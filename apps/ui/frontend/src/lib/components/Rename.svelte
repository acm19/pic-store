<script>
  import { onMount } from 'svelte';

  let directory = '';
  let newName = '';
  let isProcessing = false;
  let error = '';
  let success = false;

  let SelectDirectory, Rename;

  onMount(async () => {
    try {
      const module = await import('../wailsjs/go/main/App');
      SelectDirectory = module.SelectDirectory;
      Rename = module.Rename;
    } catch (err) {
      console.error('Failed to load Wails bindings:', err);
    }
  });

  async function selectDir() {
    try {
      const dir = await SelectDirectory();
      if (dir) directory = dir;
    } catch (err) {
      console.error('Failed to select directory:', err);
    }
  }

  async function startRename() {
    if (!directory || !newName) {
      error = 'Please select directory and enter a new name';
      return;
    }

    isProcessing = true;
    error = '';
    success = false;

    try {
      await Rename({ directory, newName });
      success = true;
    } catch (err) {
      error = err.toString();
    } finally {
      isProcessing = false;
    }
  }
</script>

<div class="rename">
  <h2>Rename Directory</h2>
  <p class="description">
    Rename a date-based directory and update all image filenames accordingly.
    Expected format: <code>YYYY MM Month DD [name]</code>
  </p>

  <div class="form">
    <div class="form-group">
      <label for="directory">Date-Based Directory</label>
      <div class="dir-input">
        <input type="text" id="directory" bind:value={directory} readonly placeholder="Select date-based directory..." />
        <button on:click={selectDir} disabled={isProcessing}>Browse</button>
      </div>
      <p class="hint">Example: 2024 01 January 15 vacation</p>
    </div>

    <div class="form-group">
      <label for="newName">New Name</label>
      <input type="text" id="newName" bind:value={newName} placeholder="beach-trip" disabled={isProcessing} />
      <p class="hint">Will rename to: 2024 01 January 15 beach-trip</p>
    </div>

    <button class="btn-primary" on:click={startRename} disabled={isProcessing || !directory || !newName}>
      {isProcessing ? 'Renaming...' : 'Rename Directory'}
    </button>
  </div>

  {#if error}
    <div class="alert alert-error">
      <strong>Error:</strong> {error}
    </div>
  {/if}

  {#if success}
    <div class="alert alert-success">
      Directory renamed successfully!
    </div>
  {/if}
</div>

<style>
  .rename {
    max-width: 800px;
  }

  h2 {
    margin: 0 0 8px 0;
    font-size: 24px;
  }

  .description {
    margin: 0 0 24px 0;
    color: var(--text-secondary);
    font-size: 14px;
  }

  .description code {
    background-color: var(--secondary-bg);
    padding: 2px 6px;
    border-radius: 3px;
    font-family: monospace;
    font-size: 13px;
  }

  .form {
    background-color: var(--secondary-bg);
    padding: 24px;
    border-radius: 8px;
    margin-bottom: 24px;
  }

  .dir-input {
    display: flex;
    gap: 8px;
  }

  .dir-input input {
    flex: 1;
  }

  .dir-input button {
    flex-shrink: 0;
  }

  .hint {
    margin: 6px 0 0 0;
    font-size: 12px;
    color: var(--text-secondary);
    font-style: italic;
  }

  .btn-primary {
    width: 100%;
    padding: 12px;
    font-size: 16px;
    margin-top: 8px;
  }

  .alert {
    padding: 16px;
    border-radius: 8px;
    margin-bottom: 16px;
  }

  .alert-error {
    background-color: rgba(244, 67, 54, 0.1);
    border: 1px solid var(--error);
    color: var(--error);
  }

  .alert-success {
    background-color: rgba(76, 175, 80, 0.1);
    border: 1px solid var(--success);
    color: var(--success);
  }
</style>
