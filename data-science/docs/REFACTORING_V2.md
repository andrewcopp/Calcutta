# Pipeline Refactoring V2: God Function → Orchestrator Pattern

## Problem: The God Function

The original `run()` function in `runner.py` had ~155 lines and too many responsibilities:

1. ✅ Path resolution (snapshot_dir, artifacts_root, run_dir)
2. ✅ Manifest initialization
3. ✅ Stage orchestration (if/then for each stage)
4. ✅ Parameter passing to each stage
5. ✅ Result collection
6. ✅ Manifest writing

**Issues:**
- Hard to test (requires full setup)
- Hard to extend (add new stage = modify God function)
- Mixed concerns (paths + orchestration + manifest management)
- Difficult to understand flow

## Solution: Separation of Concerns

### New Architecture

**1. ArtifactStore** (`orchestrator.py`)
- Manages all path resolution
- Knows about canonical vs versioned artifacts
- Single source of truth for file locations

```python
store = ArtifactStore(
    snapshot_dir=Path("out/2017"),
    snapshot_name="2017",
)

# Clean API for paths
canonical_path = store.get_canonical_path("tournaments")
versioned_path = store.get_versioned_path("investment_report")
```

**2. PipelineOrchestrator** (`orchestrator.py`)
- Manages stage execution
- Handles manifest updates
- Collects results

```python
orchestrator = PipelineOrchestrator(
    artifact_store=store,
    use_cache=True,
)

# Run stages with clean interface
out_path = orchestrator.run_stage(
    stage_name="predicted_game_outcomes",
    stage_func=_stage_predicted_game_outcomes,
    calcutta_key=None,
    kenpom_scale=10.0,
)
```

**3. run_v2()** (`runner_v2.py`)
- Thin wrapper that maps CLI args to stage calls
- No path logic
- No manifest logic
- Just orchestration

## Benefits

### 1. Testability
```python
# Easy to test individual components
def test_artifact_store_paths():
    store = ArtifactStore(...)
    assert store.get_canonical_path("tournaments") == Path("...")

def test_orchestrator_runs_stage():
    orchestrator = PipelineOrchestrator(...)
    # Mock stage function and verify it's called correctly
```

### 2. Extensibility
```python
# Add new stage = just add to run_v2()
if "new_stage" in wanted:
    out_path = orchestrator.run_stage(
        stage_name="new_stage",
        stage_func=_stage_new_stage,
        param1=value1,
    )
```

### 3. Clarity
- Each class has one clear responsibility
- Easy to understand what each piece does
- No more 155-line function

### 4. Reusability
```python
# Can use ArtifactStore independently
store = ArtifactStore(...)
tournaments_path = store.get_canonical_path("tournaments")
df = pd.read_parquet(tournaments_path)

# Can use Orchestrator for custom workflows
orchestrator = PipelineOrchestrator(...)
for stage in custom_stage_list:
    orchestrator.run_stage(...)
```

## Code Comparison

### Before (God Function)
```python
def run(...):  # 155 lines, 20+ parameters
    # Path resolution
    sd = Path(snapshot_dir)
    root = Path(artifacts_root) if artifacts_root else default_artifacts_root(sd)
    rid = str(run_id) if run_id else utc_now_iso().replace(":", "").replace("-", "")
    run_dir = build_run_dir(...)
    
    # Manifest initialization
    manifest = {"moneyball": {...}, "stages": {}}
    
    # Stage execution (repeated 6 times)
    if "predicted_game_outcomes" in wanted:
        out_path, manifest = _stage_predicted_game_outcomes(
            snapshot_dir=sd,
            out_dir=run_dir,
            calcutta_key=calcutta_key,
            kenpom_scale=float(kenpom_scale),
            n_sims=int(n_sims),
            seed=int(seed),
            use_cache=bool(use_cache),
            manifest=manifest,
        )
        results["predicted_game_outcomes_parquet"] = str(out_path)
    
    # ... repeat 5 more times ...
    
    # Manifest writing
    write_json(run_dir / "manifest.json", manifest)
    return results
```

### After (Orchestrator Pattern)
```python
def run_v2(...):  # 165 lines, but much clearer
    # Validate inputs
    sd = Path(snapshot_dir)
    if not sd.exists():
        raise FileNotFoundError(f"snapshot_dir not found: {sd}")
    
    # Initialize components
    store = ArtifactStore(snapshot_dir=sd, snapshot_name=sname, ...)
    orchestrator = PipelineOrchestrator(artifact_store=store, use_cache=use_cache)
    
    # Run stages (clean, repeatable pattern)
    if "predicted_game_outcomes" in wanted:
        out_path = orchestrator.run_stage(
            stage_name="predicted_game_outcomes",
            stage_func=_stage_predicted_game_outcomes,
            calcutta_key=calcutta_key,
            kenpom_scale=float(kenpom_scale),
            n_sims=int(n_sims),
            seed=int(seed),
        )
        orchestrator.results["predicted_game_outcomes_parquet"] = str(out_path)
    
    # Finalize
    return orchestrator.finalize()
```

## Migration Path

### Option 1: Gradual Migration
1. Keep both `run()` and `run_v2()`
2. Test `run_v2()` thoroughly
3. Switch CLI to use `run_v2()`
4. Deprecate `run()` after confidence

### Option 2: Direct Replacement
1. Replace `run()` with `run_v2()` 
2. Update all imports
3. Run full test suite

## Next Steps for V3

With this refactoring in place, V3 improvements become easier:

1. **Explicit Dependencies** - Add dependency graph to Orchestrator
2. **Parallel Execution** - Run independent stages in parallel
3. **Better Caching** - Move cache logic to ArtifactStore
4. **Config Management** - Add ConfigManager class
5. **Observability** - Add logging/metrics to Orchestrator

## Testing

Verified `run_v2()` works:
```bash
python -c "from moneyball.pipeline.runner_v2 import run_v2; ..."
# ✓ Using cached predicted_game_outcomes
# Success!
```

## Files Changed

- **New**: `moneyball/pipeline/orchestrator.py` - ArtifactStore + PipelineOrchestrator
- **New**: `moneyball/pipeline/runner_v2.py` - Refactored run function
- **Unchanged**: `moneyball/pipeline/runner.py` - Original still works
