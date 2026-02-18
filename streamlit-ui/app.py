import streamlit as st
import requests
import json
import sseclient
from typing import Optional

# Configuration
API_URL = "http://localhost:8081/api/v1"

# Page config
st.set_page_config(
    page_title="Knowledge Graph Agent Chat UI",
    page_icon="🤖",
    layout="wide",
    initial_sidebar_state="expanded"
)

# Custom CSS
st.markdown("""
<style>
    .stChatMessage {
        padding: 1rem;
        border-radius: 0.5rem;
    }
    .model-badge {
        display: inline-block;
        padding: 0.2rem 0.5rem;
        border-radius: 0.3rem;
        font-size: 0.8rem;
        font-weight: bold;
        margin-bottom: 0.5rem;
    }
    .haiku-badge {
        background-color: #10b981;
        color: white;
    }
    .sonnet-badge {
        background-color: #3b82f6;
        color: white;
    }
</style>
""", unsafe_allow_html=True)

# Initialize session state
if "messages" not in st.session_state:
    st.session_state.messages = []
if "session_id" not in st.session_state:
    st.session_state.session_id = None
if "total_tokens" not in st.session_state:
    st.session_state.total_tokens = 0

# Title
st.title("🤖 Knowledge Graph Agent Chat")
st.caption("Intelligent documentation assistant powered by Claude")

# Sidebar
with st.sidebar:
    st.header("⚙️ Settings")
    
    # Session info
    st.subheader("Session")
    if st.session_state.session_id:
        st.success(f"Active: `{st.session_state.session_id[:8]}...`")
        col1, col2 = st.columns(2)
        with col1:
            if st.button("📋 Copy ID", use_container_width=True):
                st.toast("Session ID copied!")
        with col2:
            if st.button("🔄 New", use_container_width=True):
                st.session_state.session_id = None
                st.session_state.messages = []
                st.session_state.total_tokens = 0
                st.rerun()
    else:
        st.info("No active session")
    
    st.divider()
    
    # Model parameters
    st.subheader("Parameters")
    max_tokens = st.slider("Max Tokens", 100, 2000, 500, 50)
    temperature = st.slider("Temperature", 0.0, 1.0, 0.7, 0.1)
    
    st.divider()
    
    # Streaming toggle
    st.subheader("Options")
    use_streaming = st.checkbox("Enable Streaming", value=False, help="Stream responses in real-time")
    show_metadata = st.checkbox("Show Details", value=True, help="Display model and session info")
    
    st.divider()
    
    # Stats
    st.subheader("📊 Stats")
    st.metric("Messages", len(st.session_state.messages))
    if st.session_state.total_tokens > 0:
        # Rough cost estimate (average of Haiku and Sonnet)
        avg_cost_per_1m = 1.5  # Average input cost
        estimated_cost = (st.session_state.total_tokens / 1_000_000) * avg_cost_per_1m
        st.metric("Est. Tokens", f"{st.session_state.total_tokens:,}")
        st.metric("Est. Cost", f"${estimated_cost:.4f}")
    
    st.divider()
    
    # Actions
    if st.button("🗑️ Clear Chat", use_container_width=True):
        st.session_state.messages = []
        st.session_state.total_tokens = 0
        st.rerun()
    
    # Health check
    try:
        health_response = requests.get(f"{API_URL}/health", timeout=2)
        if health_response.status_code == 200:
            st.success("✅ Agent Online")
        else:
            st.error("❌ Agent Error")
    except:
        st.error("❌ Agent Offline")

# Display chat messages
for idx, message in enumerate(st.session_state.messages):
    with st.chat_message(message["role"]):
        # Show model badge for assistant messages
        if message["role"] == "assistant" and show_metadata and "metadata" in message:
            model = message["metadata"].get("model", "")
            if "haiku" in model.lower():
                st.markdown('<div class="model-badge haiku-badge">🟢 Haiku (Fast)</div>', unsafe_allow_html=True)
            elif "sonnet" in model.lower():
                st.markdown('<div class="model-badge sonnet-badge">🔵 Sonnet (Smart)</div>', unsafe_allow_html=True)
        
        # Display message content
        st.markdown(message["content"])
        
        # Show metadata in expander
        if show_metadata and "metadata" in message and message["metadata"]:
            with st.expander("📋 Details", expanded=False):
                st.json(message["metadata"])

# Chat input
if prompt := st.chat_input("Ask me anything about the documentation..."):
    # Add user message
    st.session_state.messages.append({"role": "user", "content": prompt})
    
    # Estimate tokens for user message
    st.session_state.total_tokens += len(prompt.split()) * 1.3
    
    # Display user message
    with st.chat_message("user"):
        st.markdown(prompt)
    
    # Get assistant response
    with st.chat_message("assistant"):
        message_placeholder = st.empty()
        metadata_placeholder = st.empty()
        
        if use_streaming:
            # Streaming mode
            full_response = ""
            metadata = {}
            
            try:
                # Make streaming request
                payload = {
                    "prompt": prompt,
                    "max_tokens": max_tokens,
                    "temperature": temperature
                }
                if st.session_state.session_id:
                    payload["session_id"] = st.session_state.session_id
                
                response = requests.post(
                    f"{API_URL}/query/stream",
                    json=payload,
                    stream=True,
                    timeout=60,
                    headers={
                        "Accept": "text/event-stream",
                        "Cache-Control": "no-cache"
                    }
                )
                
                if response.status_code != 200:
                    st.error(f"API Error: {response.status_code}")
                    full_response = "Error: Could not get response from agent"
                else:
                    client = sseclient.SSEClient(response)
                    
                    for event in client.events():
                        if event.event == "start":
                            data = json.loads(event.data)
                            st.session_state.session_id = data.get("session_id")
                            metadata["model"] = data.get("model")
                            metadata["session_id"] = data.get("session_id")
                            
                            # Show model badge
                            if show_metadata:
                                model = data.get("model", "")
                                if "haiku" in model.lower():
                                    message_placeholder.markdown('<div class="model-badge haiku-badge">🟢 Haiku (Fast)</div>', unsafe_allow_html=True)
                                elif "sonnet" in model.lower():
                                    message_placeholder.markdown('<div class="model-badge sonnet-badge">🔵 Sonnet (Smart)</div>', unsafe_allow_html=True)
                        
                        elif event.event == "chunk":
                            data = json.loads(event.data)
                            chunk = data.get("text", "")
                            full_response += chunk
                            
                            # Update message with typing indicator
                            display_text = full_response + "▌"
                            if show_metadata and metadata.get("model"):
                                model = metadata["model"]
                                badge = ""
                                if "haiku" in model.lower():
                                    badge = '<div class="model-badge haiku-badge">🟢 Haiku (Fast)</div>'
                                elif "sonnet" in model.lower():
                                    badge = '<div class="model-badge sonnet-badge">🔵 Sonnet (Smart)</div>'
                                message_placeholder.markdown(f"{badge}\n\n{display_text}", unsafe_allow_html=True)
                            else:
                                message_placeholder.markdown(display_text)
                        
                        elif event.event == "done":
                            data = json.loads(event.data)
                            metadata["stop_reason"] = data.get("stop_reason")
                            
                            # Final display without cursor
                            if show_metadata and metadata.get("model"):
                                model = metadata["model"]
                                badge = ""
                                if "haiku" in model.lower():
                                    badge = '<div class="model-badge haiku-badge">🟢 Haiku (Fast)</div>'
                                elif "sonnet" in model.lower():
                                    badge = '<div class="model-badge sonnet-badge">🔵 Sonnet (Smart)</div>'
                                message_placeholder.markdown(f"{badge}\n\n{full_response}", unsafe_allow_html=True)
                            else:
                                message_placeholder.markdown(full_response)
                        
                        elif event.event == "error":
                            data = json.loads(event.data)
                            st.error(f"⚠️ Error: {data.get('error')}")
                            full_response = f"Error: {data.get('error')}"
                            break
                    
                    # Estimate tokens for response
                    st.session_state.total_tokens += len(full_response.split()) * 1.3
                    
                    # Show metadata expander
                    if show_metadata and metadata:
                        with metadata_placeholder.expander("📋 Details", expanded=False):
                            st.json(metadata)
                
            except requests.exceptions.Timeout:
                st.error("⚠️ Request timed out. Please try again.")
                full_response = "Error: Request timed out"
            except requests.exceptions.ConnectionError:
                st.error("⚠️ Could not connect to agent. Is it running?")
                full_response = "Error: Connection failed"
            except Exception as e:
                st.error(f"⚠️ Unexpected error: {str(e)}")
                full_response = f"Error: {str(e)}"
        
        else:
            # Non-streaming mode
            try:
                payload = {
                    "prompt": prompt,
                    "max_tokens": max_tokens,
                    "temperature": temperature
                }
                if st.session_state.session_id:
                    payload["session_id"] = st.session_state.session_id
                
                with st.spinner("Thinking..."):
                    response = requests.post(
                        f"{API_URL}/query",
                        json=payload,
                        timeout=60
                    )
                    response.raise_for_status()
                
                data = response.json()
                full_response = data.get("content", "No response")
                st.session_state.session_id = data.get("session_id")
                
                metadata = {
                    "session_id": data.get("session_id"),
                    "model": data.get("model"),
                    "stop_reason": data.get("stop_reason")
                }
                
                # Show model badge
                if show_metadata:
                    model = data.get("model", "")
                    badge = ""
                    if "haiku" in model.lower():
                        badge = '<div class="model-badge haiku-badge">🟢 Haiku (Fast)</div>'
                    elif "sonnet" in model.lower():
                        badge = '<div class="model-badge sonnet-badge">🔵 Sonnet (Smart)</div>'
                    message_placeholder.markdown(f"{badge}\n\n{full_response}", unsafe_allow_html=True)
                else:
                    message_placeholder.markdown(full_response)
                
                # Estimate tokens
                st.session_state.total_tokens += len(full_response.split()) * 1.3
                
                # Show metadata
                if show_metadata and metadata:
                    with metadata_placeholder.expander("📋 Details", expanded=False):
                        st.json(metadata)
                
            except requests.exceptions.Timeout:
                st.error("⚠️ Request timed out. Please try again.")
                full_response = "Error: Request timed out"
            except requests.exceptions.ConnectionError:
                st.error("⚠️ Could not connect to agent. Is it running?")
                full_response = "Error: Connection failed"
            except Exception as e:
                st.error(f"⚠️ Failed to get response: {str(e)}")
                full_response = f"Error: {str(e)}"
        
        # Add assistant message to chat history
        st.session_state.messages.append({
            "role": "assistant",
            "content": full_response,
            "metadata": metadata if 'metadata' in locals() else {}
        })

# Footer
st.divider()
col1, col2, col3 = st.columns([2, 1, 1])
with col1:
    st.caption("🤖 KG Agent - Powered by Claude 3.5")
with col2:
    st.caption("Made with Streamlit")
with col3:
    if st.button("ℹ️ About"):
        st.info("""
        **KG Agent Chat UI**
        
        This is a documentation assistant that uses:
        - 🟢 **Haiku**: Fast model for simple queries
        - 🔵 **Sonnet**: Smart model for complex queries
        - 🧠 **Smart Retrieval**: Decides when to search docs
        - 💬 **Conversation Memory**: Multi-turn conversations
        """)
