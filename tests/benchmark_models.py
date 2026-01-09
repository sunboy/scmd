#!/usr/bin/env python3
"""
Comprehensive Model Benchmarking Suite for scmd
Tests multiple lightweight models for performance, quality, and reliability
"""

import json
import os
import time
import subprocess
import statistics
import psutil
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Tuple
import tempfile

class ModelBenchmark:
    """Comprehensive model benchmarking for scmd"""

    def __init__(self, scmd_path: str = "/Users/sandeep/Projects/scmd/scmd"):
        self.scmd_path = scmd_path
        self.results = {}
        self.test_models = [
            "qwen2.5-1.5b",  # Default in code
            "qwen2.5-0.5b",  # Ultra-fast
            "qwen2.5-3b",    # Quality option
            "qwen3-4b",      # Alternative (currently default)
        ]

        # Test data
        self.test_prompts = {
            "small": "What is a variable?",
            "medium": "Explain this Python code:\ndef fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)",
            "large": """Review this code for best practices and potential issues:
import requests
import json

class APIClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()

    def get_data(self, endpoint):
        url = self.base_url + endpoint
        headers = {'Authorization': f'Bearer {self.api_key}'}
        response = self.session.get(url, headers=headers)
        return json.loads(response.text)

    def post_data(self, endpoint, data):
        url = self.base_url + endpoint
        headers = {
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json'
        }
        response = self.session.post(url, headers=headers, data=json.dumps(data))
        if response.status_code == 200:
            return json.loads(response.text)
        else:
            return None

    def __del__(self):
        self.session.close()
"""
        }

        self.code_samples = [
            ("python", "def is_prime(n):\n    return n > 1 and all(n % i != 0 for i in range(2, int(n**0.5) + 1))"),
            ("javascript", "const debounce = (func, wait) => {\n  let timeout;\n  return (...args) => {\n    clearTimeout(timeout);\n    timeout = setTimeout(() => func(...args), wait);\n  };\n};"),
            ("go", "func quickSort(arr []int) []int {\n    if len(arr) < 2 {\n        return arr\n    }\n    pivot := arr[0]\n    var less, greater []int\n    for _, v := range arr[1:] {\n        if v <= pivot {\n            less = append(less, v)\n        } else {\n            greater = append(greater, v)\n        }\n    }\n    return append(append(quickSort(less), pivot), quickSort(greater)...)\n}"),
        ]

        self.questions = [
            "What is the time complexity of binary search?",
            "How do I fix a memory leak in Node.js?",
            "What's the difference between git merge and git rebase?",
            "How can I optimize database queries in PostgreSQL?",
            "What are the SOLID principles in software design?",
        ]

    def measure_response_time(self, model: str, prompt: str, warmup: bool = False) -> Dict[str, float]:
        """Measure response time for a single query"""
        env = os.environ.copy()
        env['SCMD_MODEL'] = model
        env['SCMD_QUIET'] = '1'

        # Create temp file for prompt
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as f:
            f.write(prompt)
            prompt_file = f.name

        try:
            start_time = time.time()
            result = subprocess.run(
                [self.scmd_path, 'explain', '-'],
                input=prompt.encode(),
                capture_output=True,
                text=False,
                env=env,
                timeout=30
            )
            end_time = time.time()

            total_time = end_time - start_time
            output = result.stdout.decode('utf-8', errors='ignore')

            # Estimate TTFT (simplified - first chunk of output)
            ttft = total_time * 0.3 if output else total_time

            return {
                "total_time": total_time,
                "ttft": ttft,
                "output_length": len(output),
                "success": result.returncode == 0,
                "output": output[:500]  # First 500 chars for analysis
            }
        except subprocess.TimeoutExpired:
            return {
                "total_time": 30.0,
                "ttft": 30.0,
                "output_length": 0,
                "success": False,
                "output": "TIMEOUT"
            }
        except Exception as e:
            return {
                "total_time": -1,
                "ttft": -1,
                "output_length": 0,
                "success": False,
                "output": str(e)
            }
        finally:
            os.unlink(prompt_file)

    def benchmark_performance(self, model: str) -> Dict[str, Any]:
        """Benchmark performance metrics for a model"""
        print(f"  Testing performance for {model}...")
        results = {
            "cold_start": {},
            "warm_queries": {},
            "throughput": {}
        }

        # Cold start test
        print("    - Cold start test")
        cold_result = self.measure_response_time(model, self.test_prompts["small"])
        results["cold_start"] = {
            "time": cold_result["total_time"],
            "ttft": cold_result["ttft"],
            "success": cold_result["success"]
        }

        # Warm queries with different sizes
        print("    - Warm query tests")
        for size, prompt in self.test_prompts.items():
            times = []
            for i in range(3):  # 3 runs per size
                result = self.measure_response_time(model, prompt, warmup=True)
                if result["success"]:
                    times.append(result["total_time"])

            if times:
                results["warm_queries"][size] = {
                    "avg_time": statistics.mean(times),
                    "min_time": min(times),
                    "max_time": max(times),
                    "std_dev": statistics.stdev(times) if len(times) > 1 else 0
                }
            else:
                results["warm_queries"][size] = {"error": "All attempts failed"}

        # Throughput test (tokens per second estimate)
        print("    - Throughput test")
        prompt = self.test_prompts["medium"]
        result = self.measure_response_time(model, prompt)
        if result["success"]:
            # Rough estimate: ~4 chars per token
            input_tokens = len(prompt) / 4
            output_tokens = result["output_length"] / 4
            total_tokens = input_tokens + output_tokens

            results["throughput"] = {
                "tokens_per_second": total_tokens / result["total_time"] if result["total_time"] > 0 else 0,
                "input_tokens": input_tokens,
                "output_tokens": output_tokens,
                "time": result["total_time"]
            }

        return results

    def benchmark_quality(self, model: str) -> Dict[str, Any]:
        """Benchmark quality metrics for a model"""
        print(f"  Testing quality for {model}...")
        results = {
            "code_explanation": {},
            "code_review": {},
            "question_answering": {}
        }

        # Code explanation quality
        print("    - Code explanation tests")
        explanations = []
        for lang, code in self.code_samples[:2]:  # Test 2 samples
            result = self.measure_response_time(model, f"Explain this {lang} code:\n{code}")
            if result["success"]:
                # Simple quality scoring based on output characteristics
                score = self.score_explanation_quality(result["output"], code)
                explanations.append(score)

        if explanations:
            results["code_explanation"] = {
                "avg_score": statistics.mean(explanations),
                "samples_tested": len(explanations)
            }

        # Question answering quality
        print("    - Question answering tests")
        qa_scores = []
        for question in self.questions[:3]:  # Test 3 questions
            result = self.measure_response_time(model, question)
            if result["success"]:
                score = self.score_answer_quality(result["output"], question)
                qa_scores.append(score)

        if qa_scores:
            results["question_answering"] = {
                "avg_score": statistics.mean(qa_scores),
                "samples_tested": len(qa_scores)
            }

        return results

    def benchmark_resources(self, model: str) -> Dict[str, Any]:
        """Benchmark resource usage for a model"""
        print(f"  Testing resource usage for {model}...")
        results = {
            "memory": {},
            "cpu": {}
        }

        # Get initial system state
        process = psutil.Process()
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB

        # Run a query and monitor resources
        env = os.environ.copy()
        env['SCMD_MODEL'] = model
        env['SCMD_QUIET'] = '1'

        proc = subprocess.Popen(
            [self.scmd_path, 'explain', '-'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=env
        )

        # Monitor while running
        max_memory = initial_memory
        cpu_samples = []

        try:
            # Send input
            proc.stdin.write(self.test_prompts["medium"].encode())
            proc.stdin.close()

            # Monitor for up to 10 seconds
            start_time = time.time()
            while time.time() - start_time < 10:
                if proc.poll() is not None:
                    break

                try:
                    p = psutil.Process(proc.pid)
                    mem = p.memory_info().rss / 1024 / 1024
                    cpu = p.cpu_percent(interval=0.1)

                    max_memory = max(max_memory, mem)
                    cpu_samples.append(cpu)
                except:
                    pass

                time.sleep(0.1)

            proc.wait(timeout=5)

        except:
            proc.kill()

        results["memory"] = {
            "initial_mb": initial_memory,
            "peak_mb": max_memory,
            "delta_mb": max_memory - initial_memory
        }

        if cpu_samples:
            results["cpu"] = {
                "avg_percent": statistics.mean(cpu_samples),
                "max_percent": max(cpu_samples)
            }

        return results

    def benchmark_reliability(self, model: str) -> Dict[str, Any]:
        """Benchmark reliability metrics for a model"""
        print(f"  Testing reliability for {model}...")
        results = {
            "consistency": {},
            "edge_cases": {},
            "stability": {}
        }

        # Consistency test - same prompt multiple times
        print("    - Consistency test")
        prompt = "What is a Python decorator?"
        responses = []
        for i in range(5):
            result = self.measure_response_time(model, prompt)
            if result["success"]:
                responses.append(result["output"])

        if len(responses) >= 2:
            # Calculate similarity between responses (simplified)
            similarities = []
            for i in range(len(responses)-1):
                similarity = self.calculate_similarity(responses[i], responses[i+1])
                similarities.append(similarity)

            results["consistency"] = {
                "avg_similarity": statistics.mean(similarities),
                "runs": len(responses)
            }

        # Edge case tests
        print("    - Edge case tests")
        edge_cases = [
            ("empty", ""),
            ("special_chars", "Explain: 🚀 => λx.x²"),
            ("malformed", "def func(\n    print("),
        ]

        edge_results = {}
        for name, prompt in edge_cases:
            result = self.measure_response_time(model, prompt if prompt else " ")
            edge_results[name] = {
                "handled": result["success"],
                "time": result["total_time"] if result["success"] else -1
            }

        results["edge_cases"] = edge_results

        # Stability test (rapid queries)
        print("    - Stability test")
        success_count = 0
        total_runs = 10

        for i in range(total_runs):
            result = self.measure_response_time(model, f"Quick test {i}")
            if result["success"]:
                success_count += 1

        results["stability"] = {
            "success_rate": success_count / total_runs,
            "successful_runs": success_count,
            "total_runs": total_runs
        }

        return results

    def score_explanation_quality(self, output: str, code: str) -> float:
        """Score the quality of a code explanation (0-10)"""
        score = 5.0  # Base score

        # Check for key indicators of quality
        if len(output) > 100:
            score += 1  # Detailed explanation
        if len(output) > 300:
            score += 1  # Very detailed

        # Check for technical terms
        tech_terms = ["function", "variable", "return", "loop", "condition", "algorithm", "complexity"]
        term_count = sum(1 for term in tech_terms if term.lower() in output.lower())
        score += min(term_count * 0.5, 2)

        # Check if it mentions specific code elements
        if any(func in output for func in ["def ", "function", "method"]):
            score += 1

        return min(score, 10.0)

    def score_answer_quality(self, output: str, question: str) -> float:
        """Score the quality of an answer (0-10)"""
        score = 5.0  # Base score

        # Length check
        if len(output) > 50:
            score += 1
        if len(output) > 200:
            score += 1

        # Relevance check (simplified)
        question_words = question.lower().split()
        relevant_words = sum(1 for word in question_words if word in output.lower())
        score += min(relevant_words * 0.3, 2)

        # Structure check
        if "\n" in output or "." in output:
            score += 1  # Has structure

        return min(score, 10.0)

    def calculate_similarity(self, text1: str, text2: str) -> float:
        """Calculate similarity between two texts (0-1)"""
        # Simple word overlap similarity
        words1 = set(text1.lower().split())
        words2 = set(text2.lower().split())

        if not words1 or not words2:
            return 0.0

        intersection = words1.intersection(words2)
        union = words1.union(words2)

        return len(intersection) / len(union) if union else 0.0

    def calculate_overall_score(self, model_results: Dict) -> Dict[str, float]:
        """Calculate overall scores for a model"""
        scores = {
            "performance": 0.0,
            "quality": 0.0,
            "reliability": 0.0,
            "overall": 0.0
        }

        # Performance score (40% weight)
        perf = model_results.get("performance", {})
        if perf:
            # Response time scoring
            warm_times = []
            for size_data in perf.get("warm_queries", {}).values():
                if isinstance(size_data, dict) and "avg_time" in size_data:
                    warm_times.append(size_data["avg_time"])

            if warm_times:
                avg_time = statistics.mean(warm_times)
                # Score: 10 points for <1s, decreasing linearly
                time_score = max(0, 10 - (avg_time - 1) * 2)
                scores["performance"] = min(time_score, 10)

        # Quality score (40% weight)
        qual = model_results.get("quality", {})
        if qual:
            quality_scores = []

            if "code_explanation" in qual and "avg_score" in qual["code_explanation"]:
                quality_scores.append(qual["code_explanation"]["avg_score"])

            if "question_answering" in qual and "avg_score" in qual["question_answering"]:
                quality_scores.append(qual["question_answering"]["avg_score"])

            if quality_scores:
                scores["quality"] = statistics.mean(quality_scores)

        # Reliability score (20% weight)
        rel = model_results.get("reliability", {})
        if rel:
            reliability_scores = []

            if "consistency" in rel and "avg_similarity" in rel["consistency"]:
                reliability_scores.append(rel["consistency"]["avg_similarity"] * 10)

            if "stability" in rel and "success_rate" in rel["stability"]:
                reliability_scores.append(rel["stability"]["success_rate"] * 10)

            if reliability_scores:
                scores["reliability"] = statistics.mean(reliability_scores)

        # Overall score
        scores["overall"] = (
            scores["performance"] * 0.4 +
            scores["quality"] * 0.4 +
            scores["reliability"] * 0.2
        )

        return scores

    def run_full_benchmark(self) -> Dict[str, Any]:
        """Run complete benchmark suite for all models"""
        print("\n" + "="*60)
        print("scmd Model Benchmark Suite")
        print("="*60 + "\n")

        all_results = {}

        for model in self.test_models:
            print(f"\nBenchmarking model: {model}")
            print("-" * 40)

            model_results = {}

            try:
                # Run all benchmark categories
                model_results["performance"] = self.benchmark_performance(model)
                model_results["quality"] = self.benchmark_quality(model)
                model_results["resources"] = self.benchmark_resources(model)
                model_results["reliability"] = self.benchmark_reliability(model)

                # Calculate scores
                model_results["scores"] = self.calculate_overall_score(model_results)

                # Determine tier
                overall_score = model_results["scores"]["overall"]
                if overall_score >= 9:
                    tier = "S"
                elif overall_score >= 8:
                    tier = "A"
                elif overall_score >= 7:
                    tier = "B"
                elif overall_score >= 6:
                    tier = "C"
                else:
                    tier = "D"

                model_results["tier"] = tier

                print(f"\n  Overall Score: {overall_score:.2f}/10 (Tier {tier})")

            except Exception as e:
                print(f"  ERROR: Failed to benchmark {model}: {e}")
                model_results["error"] = str(e)

            all_results[model] = model_results

        # Generate comparison matrix
        comparison = self.generate_comparison_matrix(all_results)

        # Find best model
        best_model = self.find_best_model(all_results)

        return {
            "timestamp": datetime.now().isoformat(),
            "models": all_results,
            "comparison": comparison,
            "recommendation": best_model
        }

    def generate_comparison_matrix(self, results: Dict) -> Dict:
        """Generate a comparison matrix of all models"""
        matrix = {
            "response_times": {},
            "quality_scores": {},
            "resource_usage": {},
            "overall_scores": {}
        }

        for model, data in results.items():
            if "error" in data:
                continue

            # Response times
            if "performance" in data and "warm_queries" in data["performance"]:
                times = []
                for size_data in data["performance"]["warm_queries"].values():
                    if isinstance(size_data, dict) and "avg_time" in size_data:
                        times.append(size_data["avg_time"])

                if times:
                    matrix["response_times"][model] = statistics.mean(times)

            # Quality scores
            if "scores" in data:
                matrix["quality_scores"][model] = data["scores"].get("quality", 0)
                matrix["overall_scores"][model] = data["scores"].get("overall", 0)

            # Resource usage
            if "resources" in data and "memory" in data["resources"]:
                matrix["resource_usage"][model] = data["resources"]["memory"].get("peak_mb", 0)

        return matrix

    def find_best_model(self, results: Dict) -> Dict:
        """Find the best model based on criteria"""
        candidates = []

        for model, data in results.items():
            if "error" in data or "scores" not in data:
                continue

            # Check if meets minimum criteria
            meets_criteria = True
            criteria_notes = []

            # Response time < 2s
            if "performance" in data and "warm_queries" in data["performance"]:
                times = []
                for size_data in data["performance"]["warm_queries"].values():
                    if isinstance(size_data, dict) and "avg_time" in size_data:
                        times.append(size_data["avg_time"])

                if times:
                    avg_time = statistics.mean(times)
                    if avg_time > 2:
                        criteria_notes.append(f"Response time {avg_time:.2f}s > 2s target")
                        meets_criteria = avg_time < 3  # Allow up to 3s

            # Quality score > 70%
            quality_score = data["scores"].get("quality", 0) * 10
            if quality_score < 70:
                criteria_notes.append(f"Quality score {quality_score:.0f}% < 70% target")
                meets_criteria = meets_criteria and quality_score > 60

            # Memory < 4GB
            if "resources" in data and "memory" in data["resources"]:
                peak_mb = data["resources"]["memory"].get("peak_mb", 0)
                if peak_mb > 4096:
                    criteria_notes.append(f"Memory {peak_mb:.0f}MB > 4GB limit")
                    meets_criteria = False

            candidates.append({
                "model": model,
                "overall_score": data["scores"]["overall"],
                "tier": data.get("tier", "?"),
                "meets_criteria": meets_criteria,
                "notes": criteria_notes
            })

        # Sort by overall score
        candidates.sort(key=lambda x: x["overall_score"], reverse=True)

        # Find best that meets criteria
        best = None
        for candidate in candidates:
            if candidate["meets_criteria"]:
                best = candidate
                break

        # If none meet criteria, take the best overall
        if not best and candidates:
            best = candidates[0]

        return best

    def save_results(self, results: Dict, filename: str = "benchmark_results.json"):
        """Save benchmark results to file"""
        output_path = Path(f"/Users/sandeep/Projects/scmd/{filename}")

        with open(output_path, 'w') as f:
            json.dump(results, f, indent=2)

        print(f"\nResults saved to: {output_path}")

        # Also create a summary report
        summary_path = output_path.with_suffix('.md')
        self.generate_summary_report(results, summary_path)

    def generate_summary_report(self, results: Dict, output_path: Path):
        """Generate a markdown summary report"""
        report = []
        report.append("# scmd Model Benchmark Results")
        report.append(f"\nTimestamp: {results['timestamp']}")
        report.append("\n## Executive Summary")

        if results.get("recommendation"):
            rec = results["recommendation"]
            report.append(f"\n**Recommended Model**: {rec['model']} (Tier {rec['tier']}, Score: {rec['overall_score']:.2f}/10)")

            if rec.get("notes"):
                report.append("\n**Notes**:")
                for note in rec["notes"]:
                    report.append(f"- {note}")

        report.append("\n## Model Comparison")
        report.append("\n| Model | Tier | Overall | Performance | Quality | Reliability | Avg Response Time | Memory |")
        report.append("|-------|------|---------|-------------|---------|-------------|-------------------|--------|")

        for model, data in results["models"].items():
            if "scores" in data:
                scores = data["scores"]
                tier = data.get("tier", "?")

                # Get avg response time
                avg_time = "N/A"
                if model in results["comparison"]["response_times"]:
                    avg_time = f"{results['comparison']['response_times'][model]:.2f}s"

                # Get memory usage
                memory = "N/A"
                if model in results["comparison"]["resource_usage"]:
                    memory = f"{results['comparison']['resource_usage'][model]:.0f}MB"

                report.append(
                    f"| {model} | {tier} | {scores['overall']:.2f} | "
                    f"{scores['performance']:.2f} | {scores['quality']:.2f} | "
                    f"{scores['reliability']:.2f} | {avg_time} | {memory} |"
                )

        report.append("\n## Detailed Results")

        for model, data in results["models"].items():
            report.append(f"\n### {model}")

            if "error" in data:
                report.append(f"\n**Error**: {data['error']}")
                continue

            # Performance details
            if "performance" in data and "warm_queries" in data["performance"]:
                report.append("\n**Performance**:")
                for size, metrics in data["performance"]["warm_queries"].items():
                    if isinstance(metrics, dict) and "avg_time" in metrics:
                        report.append(f"- {size}: {metrics['avg_time']:.2f}s avg")

            # Quality details
            if "quality" in data:
                report.append("\n**Quality**:")
                if "code_explanation" in data["quality"] and "avg_score" in data["quality"]["code_explanation"]:
                    report.append(f"- Code Explanation: {data['quality']['code_explanation']['avg_score']:.1f}/10")
                if "question_answering" in data["quality"] and "avg_score" in data["quality"]["question_answering"]:
                    report.append(f"- Q&A: {data['quality']['question_answering']['avg_score']:.1f}/10")

            # Reliability details
            if "reliability" in data:
                report.append("\n**Reliability**:")
                if "stability" in data["reliability"] and "success_rate" in data["reliability"]["stability"]:
                    report.append(f"- Stability: {data['reliability']['stability']['success_rate']*100:.0f}% success rate")
                if "consistency" in data["reliability"] and "avg_similarity" in data["reliability"]["consistency"]:
                    report.append(f"- Consistency: {data['reliability']['consistency']['avg_similarity']:.2f} similarity")

        with open(output_path, 'w') as f:
            f.write('\n'.join(report))

        print(f"Summary report saved to: {output_path}")


def main():
    """Run the complete benchmark suite"""
    benchmark = ModelBenchmark()

    # Check if scmd exists
    if not Path(benchmark.scmd_path).exists():
        print(f"ERROR: scmd not found at {benchmark.scmd_path}")
        print("Please build scmd first with: go build")
        return 1

    # Run benchmarks
    results = benchmark.run_full_benchmark()

    # Save results
    benchmark.save_results(results)

    # Print summary
    print("\n" + "="*60)
    print("Benchmark Complete!")
    print("="*60)

    if results.get("recommendation"):
        rec = results["recommendation"]
        print(f"\nRecommended Model: {rec['model']}")
        print(f"Overall Score: {rec['overall_score']:.2f}/10 (Tier {rec['tier']})")

        if rec.get("notes"):
            print("\nConsiderations:")
            for note in rec["notes"]:
                print(f"  - {note}")

    return 0


if __name__ == "__main__":
    sys.exit(main())