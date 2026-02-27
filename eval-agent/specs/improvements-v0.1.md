# Eval Agent Improvements

## Overview

This document outlines the phased implementation plan to evolve eval-agent from a **real-time evaluation API** into a comprehensive **evaluation platform** that supports both production monitoring and research-grade validation workflows.

**Current State:**
- Real-time HTTP API for single-request evaluation
- Two-stage pipeline (prechecks + 5 LLM judges)
- Early-exit optimization for cost savings
- MCP integration for Claude Code
- Redis Stream support for async processing

**Target State:**
- ✅ YAML-driven configurable judges **(COMPLETED)**
- Batch dataset evaluation CLI
- Human annotation validation workflow
- Correlation analysis (Kendall's tau)
- Iterative prompt improvement loop
- Reference-based vs reference-free modes
