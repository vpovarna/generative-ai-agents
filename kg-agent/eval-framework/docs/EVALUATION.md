# Search Evaluation Specification

## Overview

This document outlines the evaluation plan for testing the KG Agent's search quality using the Natural Questions dataset. The evaluation focuses on **information retrieval performance** (search accuracy) rather than end-to-end answer quality.

**Dataset:** Natural Questions (Filtered)
- **Size:** 86,212 questions with long/short answers
- **Source:** All questions from Wikipedia articles
- **Structure:** Question + Long Answer + Short Answer

**Goal:** Measure if the hybrid search system can accurately retrieve the correct answer chunk for a given question.

---

## Data Structure

### Input Format (Natural Questions CSV)

```csv
question,long_answer,short_answer
"which is the most common use of opt-in e-mail marketing","A common example of permission marketing is a newsletter sent to an advertising firm's customers. Such newsletters inform customers of upcoming events or promotions, or new products...","A newsletter sent to an advertising firm's customers"
```

### Prepared JSON Format

Transform each row into:

```json
[
  {
    "chunk_id": "nq_000000",
    "document_id": "natural_questions",
    "content": "`` Don't You (Forget About Me) '' is a 1985 pop song performed by Scottish rock band Simple Minds. The song is best known for being played during the opening and closing credits of the John Hughes film The Breakfast Club. It was written and composed by producer Keith Dorsey and Steve Schiff, the latter of whom was a guitarist and songwriter from the Nina Hagen band.",
    "metadata": {
      "question": "what film has the song don't you forget about me",
      "short_answer": "The Breakfast Club"
    }
  }
]
```

**Field Descriptions:**
- `chunk_id`: Unique identifier for the chunk (format: `nq_XXXXXX`)
- `document_id`: Source document (all use `"natural_questions"`)
- `content`: The long answer text (what gets embedded and searched)
- `metadata.question`: Original question (for debugging and reference)
- `metadata.short_answer`: Short answer (for future LLM judge evaluation)

---

### Evaluation

**Script:** `scripts/eval_search.py`

### Test Plan
### Step 1 (Evaluation):
  Insert:  1,000 chunks  (no index)  
  Queries: 200 questions (random sample with seed)  
  Goal:    Establish baseline Recall@1, Recall@5, MRR  

#### Semantic Search Results:
```json
{
    "semantic":
    {
        "recall_at_1": 0.88,
        "recall_at_5": 0.955,
        "mrr": 0.9158333333333333,
        "precision": 0.9158333333333333,
        "f1_score": 0.8975591647331785
    }
}
```

#### Notes:
**Recall@1** = 0.88 → 88% of the time, the correct chunk is the very first result. For a RAG system, this is ideal, the LLM gets the right context immediately without needing to process noise from lower-ranked results.  

**Recall@5** = 0.955 → Only 4.5% of questions (about 9 out of 200) completely failed to retrieve the correct chunk in the top 5. These are your hard failure cases worth investigating.  

**MRR** ≈ 0.916 means the average correct rank is between 1 and 2:
  rank 1 → 1/1 = 1.000
  rank 2 → 1/2 = 0.500
  rank 3 → 1/3 = 0.333

A score of 0.916 means most results land at rank 1,
with very few pushing into rank 2 or 3.

#### Keyword Search Results:
```json
{
    "keyword":
    {
        "recall_at_1": 0.125,
        "recall_at_5": 0.13,
        "mrr": 0.1275,
        "precision": 0.1275,
        "f1_score": 0.12623762376237624
    }
}
```

#### Notes:
The values are terrible. This is a textbook IR failure. Keyword search (PostgreSQL tsvector / plainto_tsquery) looks for shared words between the query and the content. Your data has an extreme mismatch:

```json
Query (question):  "what film has the song dont you forget about me"
                   After stop-word removal: "film song dont forget"

Content (answer):  "Don't You (Forget About Me) is a 1985 pop song
                   performed by Scottish rock band Simple Minds.
                   The song is best known for being played during
                   the opening and closing credits of The Breakfast Club."
```

Keyword search is binary on this dataset: either it finds a strong term match and ranks it first, or it finds nothing useful at all. It's not producing ranked candidates, it's producing near-zero results with rare correct hits.

TODO: test with BM25 instead of tsvector.

#### Hybrid Search Results:
```json
{
    "hybrid":
    {
        "recall_at_1": 0.87,
        "recall_at_5": 0.955,
        "mrr": 0.9108333333333333,
        "precision": 0.9108333333333333,
        "f1_score": 0.8899485259709873
    }
}
```

#### Notes:
Hybrid is marginally worse than pure semantic, not better. This is expected given the keyword results.
Since keyword search only finds correct answers 13% of the time, the other 87% it's injecting wrong chunks into the RRF pool — pushing some correct semantic results from rank 1 to rank 2. That's where the −0.010 on Recall@1 comes from.

On this dataset, keyword search is net negative for hybrid fusion. The RRF gets more signal from wrong results than right ones.


### Step 2:
  Insert:  10,000 chunks  (add ivfflat, lists=100, probes=1)  
  Queries: 500 questions  
  Goal:    See if metrics degrade → shows real search difficulty with index.   

#### Semantic Search Results:

```json
{
    "semantic":
    {
        "recall_at_1": 0.18,
        "recall_at_5": 0.196,
        "mrr": 0.18706666666666666,
        "precision": 0.18706666666666666,
        "f1_score": 0.18346531057028698
    }
}
```

####  Notes:
With lists=100 and the default probes=1, 99% of your data is never touched. The correct chunk is almost always in a different cluster than the one being probed.

### Step 3:
  Insert:  10,000 chunks  (add ivfflat, lists=100, probes=10)  
  Queries: 500 questions  
  Goal:    See if metrics degrade → shows real search difficulty with index.   

Updating the index probes:
`SET ivfflat.probes = 10;`

#### Semantic Search Results:
```json
{
    "semantic":
    {
        "recall_at_1": 0.718,
        "recall_at_5": 0.9,
        "mrr": 0.797,
        "precision": 0.797,
        "f1_score": 0.7554402640264026
    }
}
```
#### Notes:
The probes fix was the entire problem. Going from 1% to 10% of the corpus searched recovered almost all the accuracy.
The drop from 1K → 10K is expected and healthy:
Recall@1: 0.880 → 0.718 (−16%) — 10x more candidates competing, correct chunk harder to rank #1
Recall@5: 0.955 → 0.900 (−5.5%) — still meeting the >0.90 target
MRR: 0.916 → 0.797 — correct answer now lands at rank 1–2 on average instead of almost always rank 1

### Step 4:
  Insert:  10,000 chunks  (add hnsw, m = 16, ef_construction = 64)  
  Queries: 500 questions  
  Goal:    See if metrics degrade → shows real search difficulty with index.  

Setting the index:
```
CREATE INDEX idx_chunks_embedding 
ON document_chunks 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

#### Semantic Search Results:
```json
{
    "semantic":
    {
        "recall_at_1": 0.718,
        "recall_at_5": 0.9,
        "mrr": 0.797,
        "precision": 0.797,
        "f1_score": 0.7554402640264026
    }
}
```
#### Notes:
The index is no longer the bottleneck. hnsw gives near-sequential-scan accuracy, it matches ivfflat probes=10.


### Step 4:
  Insert:  10,000 chunks, no idex  
  Queries: 500 questions  
  Goal:    See if metrics improve without an index  

#### Semantic Search Results:
```json
{
    "semantic":
    {
        "recall_at_1": 0.718,
        "recall_at_5": 0.9,
        "mrr": 0.797,
        "precision": 0.797,
        "f1_score": 0.7554402640264026
    }
}
```

### Notes:
Removing index didn't help improve the scores. 

### Summary
Step 1 — 1K corpus, no index (baseline)
Semantic search performed excellently (Recall@1: 0.88, Recall@5: 0.955, MRR: 0.916), but these numbers are optimistic — with only 1K candidates there's minimal competition. Keyword search completely failed (Recall@1: 0.125) due to vocabulary mismatch between questions and answers, and hybrid was marginally worse than pure semantic because keyword injected wrong chunks into the RRF pool.

Step 2 — 10K corpus, ivfflat probes=1 (misconfigured index)
Metrics collapsed to Recall@1: 0.18 — a false disaster caused entirely by the default probes=1 setting, which meant only 1% of the corpus (100 vectors out of 10K) was ever searched per query.

Step 3 — 10K corpus, ivfflat probes=10 (fixed index)
Metrics recovered significantly (Recall@1: 0.718, Recall@5: 0.900, MRR: 0.797). The ~16% drop from the 1K baseline is real and expected — 10x more candidates compete for the top 5 slots. Recall@5 still meets the >0.90 target.

Step 4 — 10K corpus, hnsw / no index
Both tests produced identical results to Step 3 (0.718/0.900/0.797), confirming that the index type is no longer the bottleneck. These numbers represent the true accuracy ceiling of the current setup. The remaining gap is the Q&A embedding asymmetry: questions and long-answer content live in different regions of Titan's embedding space. Fixing this requires embedding question + answer together as a single content field, or switching to a model with asymmetric search support.

---

## References

- Natural Questions Dataset: https://ai.google.com/research/NaturalQuestions
- Mean Reciprocal Rank: https://en.wikipedia.org/wiki/Mean_reciprocal_rank
- Information Retrieval Metrics: https://www.evidentlyai.com/ranking-metrics
- RRF (Reciprocal Rank Fusion): https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf
