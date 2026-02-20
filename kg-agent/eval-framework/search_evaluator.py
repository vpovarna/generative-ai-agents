
import random
import requests
from typing import List, Dict
import json


RANDOM_SEED = 42


class SearchEvaluator:
    def __init__(self, search_api_url: str, test_data_path: str, sample_size: int = 1000):
        self.search_api_url = search_api_url
        self.sample_size = sample_size

        with open(test_data_path, 'r') as f:
            all_data = json.load(f)
            rng = random.Random(RANDOM_SEED)
            self.test_data = rng.sample(all_data, min(sample_size, len(all_data)))

        print(f"Loaded {len(self.test_data)} test cases (seed={RANDOM_SEED})")

    def run_evaluation(self):
        search_types = ['semantic']

        all_results = {}

        for search_type in search_types:
            print(f"Evaluating {search_type.upper()} search...")
            print(f"{'='*60}")

            results = self.evaluate_search(search_type, limit=5)

            # Recall@5 = (queries where correct chunk is in top 5) / (total queries). 
            # Ex: 3 found / 5 total == 60%.
            # Obs: Empty results are marked as failure
            recall_at_1 = self.calculate_recall_at_1(results=results)
            recall_at_5 = self.calculate_recall_at_5(results=results)

            # Mean Reciprocal Rank
            mrr = self.calculate_mrr(results=results)

            # Precision@5 = (# relevant docs in results) / (total docs returned)
            # Each question has exactly 1 correct answer. Max returned responses is 5. So the max precision = 1/5 = 20%
            precision_at_5 = self.calculate_precision(results=results)

            # F1 score
            f1_score = self.calculate_f1_score(recall=recall_at_1, precision=precision_at_5)

            all_results[search_type] = {
                'recall_at_1': recall_at_1,
                'recall_at_5': recall_at_5,
                'mrr': mrr,
                'precision': precision_at_5,
                'f1_score': f1_score
            }

        print(all_results)

    def evaluate_search(self, search_type: str, limit: int = 5) -> List[Dict]:

        results = []

        for i, qa in enumerate(self.test_data):
            question = qa['metadata']['question']
            expected_chunk_id = qa['chunk_id']

            search_results = []
            try:
                response = requests.post(
                    url=f'{self.search_api_url}/{search_type}',
                    json={'query': question, 'limit': limit},
                    timeout=10
                )

                response.raise_for_status()
                response_results = response.json()['result']
                if response_results is None:
                    response_results=[]
                search_results = response_results
            except Exception as e:
                print(f"Error on question {i}: {e}")
            
            found = False
            rank = -1
            returned_ids = []

            for j, result in enumerate(search_results):
                returned_chunk_id = result.get('metadata', {}).get('chunk_id', 'unknown')
                returned_ids.append(returned_chunk_id)
                if returned_chunk_id == expected_chunk_id:
                    found = True
                    rank = j + 1    
                    break

            results.append({
                'question_id': question,
                'expected_chunk_id': expected_chunk_id,
                'found': found,
                'rank': rank,
                'returned_chunk_ids': returned_ids
            })

        return results

    def calculate_recall_at_1(self, results: dict) -> float:
        """What % of queries get the right answer first?"""
        if len(results) == 0:
            return 0.0
        return sum(1 for r in results if r['rank'] == 1) / len(results)

    def calculate_recall_at_5(self, results: dict) -> float:
        """ What % of queries find the right answer at all? """
        if len(results) == 0:
            return 0.0

        return sum(1 for r in results if r['found']) / len(results)

    def calculate_mrr(self, results: dict) -> float:
        """On average, how highly ranked is the correct answer?"""

        if len(results) == 0:
            return 0.0

        return sum(1/r['rank'] for r in results if r['rank'] > 0) / len(results)

    def calculate_f1_score(self, recall: float, precision: float) -> float:
        """F1 score is the harmonic mean of precision and recall"""
        if recall == 0 or precision == 0:
            return 0.0

        return 2 * (precision * recall) / (precision + recall)

    def calculate_precision(self, results: dict) -> float:
        """Each question has exactly 1 correct answer"""
        if len(results) == 0:
            return 0.0

        precisions = []
        for r in results:
            if (len(r['returned_chunk_ids']) > 0):
                if not r['found']:
                    precision = 0.0
                else:
                    precision = 1.0 / len(r['returned_chunk_ids'])
            else:
                precision = 0.0
            precisions.append(precision)
        avg_precision = sum(precisions) / len(precisions)

        return(avg_precision)

if __name__ == "__main__":
    search_evaluator = SearchEvaluator(
        search_api_url = 'http://localhost:8082/search/v1',
        test_data_path = '.data/natural_questions_prepared.json',
        sample_size = 500
    )

    search_evaluator.run_evaluation()