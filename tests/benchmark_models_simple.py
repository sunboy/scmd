#!/usr/bin/env python3
"""
Simplified Model Benchmarking Suite for scmd
Tests multiple lightweight models for performance and quality
"""

import json
import os
import time
import subprocess
import statistics
import sys
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any
import tempfile

class SimpleBenchmark:
    """Simplified model benchmarking for scmd"""

    def __init__(self, scmd_path: str = "/Users/sandeep/Projects/scmd/scmd"):
        self.scmd_path = scmd_path
        self.results = {}

        # Test just the currently available models
        self.test_models = [
            "qwen3-4b",      # Current default in code
        ]

        # Quick test prompts
        self.test_prompts = {
            "small": "What is a variable?",
            "medium": "Explain this code: def factorial(n): return 1 if n <= 1 else n * factorial(n-1)",
            "large": "Review this code for issues:\ndef process_data(data):\n    result = []\n    for item in data:\n        if item > 0:\n            result.append(item * 2)\n    return result"
        }

    def run_scmd_command(self, prompt: str, model: str = None) -> Dict:
        """Run a single scmd command and measure performance"""
        env = os.environ.copy()
        if model:
            env['SCMD_MODEL'] = model
        env['SCMD_QUIET'] = '1'
        env['SCMD_NO_AUTOSTART'] = '1'  # Don't auto-start server

        start_time = time.time()

        try:
            result = subprocess.run(
                [self.scmd_path, 'explain', '-'],
                input=prompt.encode(),
                capture_output=True,
                text=False,
                env=env,
                timeout=30
            )
            end_time = time.time()

            output = result.stdout.decode('utf-8', errors='ignore')
            error = result.stderr.decode('utf-8', errors='ignore')

            return {
                "success": result.returncode == 0,
                "time": end_time - start_time,
                "output": output,
                "error": error,
                "output_length": len(output)
            }
        except subprocess.TimeoutExpired:
            return {
                "success": False,
                "time": 30.0,
                "output": "",
                "error": "TIMEOUT",
                "output_length": 0
            }
        except Exception as e:
            return {
                "success": False,
                "time": -1,
                "output": "",
                "error": str(e),
                "output_length": 0
            }

    def test_model_performance(self, model: str) -> Dict:
        """Test performance of a single model"""
        print(f"\nTesting model: {model}")
        print("-" * 40)

        results = {
            "model": model,
            "tests": {},
            "metrics": {}
        }

        # Test different prompt sizes
        for size, prompt in self.test_prompts.items():
            print(f"  Testing {size} prompt...")

            # Run 3 times and average
            times = []
            success_count = 0
            outputs = []

            for i in range(3):
                result = self.run_scmd_command(prompt, model)

                if result["success"]:
                    times.append(result["time"])
                    success_count += 1
                    outputs.append(result["output"])
                    print(f"    Run {i+1}: {result['time']:.2f}s")
                else:
                    print(f"    Run {i+1}: FAILED - {result['error']}")

            if times:
                results["tests"][size] = {
                    "avg_time": statistics.mean(times),
                    "min_time": min(times),
                    "max_time": max(times),
                    "success_rate": success_count / 3,
                    "avg_output_length": statistics.mean([len(o) for o in outputs]) if outputs else 0
                }
            else:
                results["tests"][size] = {
                    "error": "All attempts failed",
                    "success_rate": 0
                }

        # Calculate overall metrics
        all_times = []
        for test_data in results["tests"].values():
            if isinstance(test_data, dict) and "avg_time" in test_data:
                all_times.append(test_data["avg_time"])

        if all_times:
            results["metrics"] = {
                "overall_avg_time": statistics.mean(all_times),
                "meets_2s_target": statistics.mean(all_times) < 2.0,
                "meets_3s_target": statistics.mean(all_times) < 3.0
            }

        return results

    def run_benchmark(self) -> Dict:
        """Run simplified benchmark"""
        print("\n" + "="*60)
        print("scmd Model Performance Test")
        print("="*60)

        # First check if scmd exists and works
        print("\nChecking scmd availability...")
        test_result = self.run_scmd_command("test", None)

        if not Path(self.scmd_path).exists():
            print(f"ERROR: scmd not found at {self.scmd_path}")
            return {"error": "scmd not found"}

        all_results = {
            "timestamp": datetime.now().isoformat(),
            "models": {}
        }

        for model in self.test_models:
            model_results = self.test_model_performance(model)
            all_results["models"][model] = model_results

        # Generate summary
        all_results["summary"] = self.generate_summary(all_results["models"])

        return all_results

    def generate_summary(self, models: Dict) -> Dict:
        """Generate summary of results"""
        summary = {
            "best_model": None,
            "performance_comparison": {}
        }

        for model_name, data in models.items():
            if "metrics" in data and "overall_avg_time" in data["metrics"]:
                summary["performance_comparison"][model_name] = {
                    "avg_response_time": data["metrics"]["overall_avg_time"],
                    "meets_2s_target": data["metrics"]["meets_2s_target"],
                    "meets_3s_target": data["metrics"]["meets_3s_target"]
                }

        # Find best model
        if summary["performance_comparison"]:
            best = min(summary["performance_comparison"].items(),
                      key=lambda x: x[1]["avg_response_time"])
            summary["best_model"] = {
                "name": best[0],
                "avg_time": best[1]["avg_response_time"]
            }

        return summary

    def save_results(self, results: Dict):
        """Save results to file"""
        output_path = Path("/Users/sandeep/Projects/scmd/benchmark_results_simple.json")

        with open(output_path, 'w') as f:
            json.dump(results, f, indent=2)

        print(f"\nResults saved to: {output_path}")

        # Print summary
        print("\n" + "="*60)
        print("SUMMARY")
        print("="*60)

        if "summary" in results and "best_model" in results["summary"]:
            best = results["summary"]["best_model"]
            print(f"\nBest Model: {best['name']}")
            print(f"Average Response Time: {best['avg_time']:.2f} seconds")

            if best['avg_time'] < 2:
                print("✅ Meets <2s target!")
            elif best['avg_time'] < 3:
                print("⚠️  Meets <3s target (acceptable)")
            else:
                print("❌ Does not meet performance targets")

        # Print detailed results
        print("\nDetailed Results:")
        for model_name, model_data in results.get("models", {}).items():
            print(f"\n{model_name}:")
            for size, test_data in model_data.get("tests", {}).items():
                if "avg_time" in test_data:
                    print(f"  {size}: {test_data['avg_time']:.2f}s (success rate: {test_data['success_rate']*100:.0f}%)")
                else:
                    print(f"  {size}: FAILED")

def main():
    """Run the simplified benchmark"""
    benchmark = SimpleBenchmark()

    # Check if scmd exists
    if not Path(benchmark.scmd_path).exists():
        print(f"ERROR: scmd not found at {benchmark.scmd_path}")
        print("Please build scmd first with: go build")
        return 1

    # Run benchmark
    results = benchmark.run_benchmark()

    # Save results
    if "error" not in results:
        benchmark.save_results(results)

    return 0

if __name__ == "__main__":
    sys.exit(main())