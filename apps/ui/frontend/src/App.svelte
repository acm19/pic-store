<script>
  import { onMount } from 'svelte';
  import Parse from './lib/components/Parse.svelte';
  import Rename from './lib/components/Rename.svelte';
  import Backup from './lib/components/Backup.svelte';
  import Restore from './lib/components/Restore.svelte';

  let activeTab = 'parse';
  let version = 'dev';

  const tabs = [
    { id: 'parse', label: 'Parse & Organise' },
    { id: 'rename', label: 'Rename Directory' },
    { id: 'backup', label: 'Backup to S3' },
    { id: 'restore', label: 'Restore from S3' },
  ];

  onMount(async () => {
    try {
      // Import Wails runtime dynamically (will be generated after first build)
      const { GetVersion } = await import('./lib/wailsjs/go/main/App');
      version = await GetVersion();
    } catch (err) {
      console.log('Wails runtime not yet available:', err);
    }
  });
</script>

<div class="app">
  <header class="header">
    <h1>Pics - Media Organiser</h1>
    <span class="version">v{version}</span>
  </header>

  <nav class="tabs">
    {#each tabs as tab}
      <button
        class="tab"
        class:active={activeTab === tab.id}
        on:click={() => activeTab = tab.id}
      >
        {tab.label}
      </button>
    {/each}
  </nav>

  <main class="content">
    {#if activeTab === 'parse'}
      <Parse />
    {:else if activeTab === 'rename'}
      <Rename />
    {:else if activeTab === 'backup'}
      <Backup />
    {:else if activeTab === 'restore'}
      <Restore />
    {/if}
  </main>
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 100%;
  }

  .header {
    background-color: var(--secondary-bg);
    padding: 16px 24px;
    border-bottom: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .header h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .version {
    font-size: 12px;
    color: var(--text-secondary);
    font-family: monospace;
  }

  .tabs {
    display: flex;
    background-color: var(--secondary-bg);
    border-bottom: 1px solid var(--border);
    padding: 0 16px;
  }

  .tab {
    background: none;
    border: none;
    border-radius: 0;
    padding: 12px 20px;
    color: var(--text-secondary);
    font-size: 14px;
    cursor: pointer;
    position: relative;
    transition: color 0.2s;
  }

  .tab:hover {
    background-color: transparent;
    color: var(--text-primary);
  }

  .tab.active {
    color: var(--accent);
  }

  .tab.active::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 2px;
    background-color: var(--accent);
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }
</style>
