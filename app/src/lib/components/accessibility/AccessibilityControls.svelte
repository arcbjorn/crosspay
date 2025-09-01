<script lang="ts">
  import { onMount } from 'svelte';
  import CyberButton from '../cyber/CyberButton.svelte';
  import CyberModal from '../cyber/CyberModal.svelte';
  
  let showControls = false;
  let settings = {
    reducedMotion: false,
    highContrast: false,
    largerText: false,
    screenReader: false,
    keyboardNav: true
  };
  
  onMount(() => {
    // Load accessibility settings
    const saved = localStorage.getItem('crossPayAccessibility');
    if (saved) {
      try {
        settings = { ...settings, ...JSON.parse(saved) };
        applySettings();
      } catch (e) {
        console.error('Failed to load accessibility settings:', e);
      }
    }
    
    // Check system preferences
    checkSystemPreferences();
  });
  
  function checkSystemPreferences() {
    // Reduced motion
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      settings.reducedMotion = true;
    }
    
    // High contrast
    if (window.matchMedia('(prefers-contrast: high)').matches) {
      settings.highContrast = true;
    }
    
    // Large text
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      // Most users prefer dark mode, already implemented
    }
    
    applySettings();
  }
  
  function applySettings() {
    const root = document.documentElement;
    
    // Reduced motion
    if (settings.reducedMotion) {
      root.classList.add('reduce-motion');
    } else {
      root.classList.remove('reduce-motion');
    }
    
    // High contrast
    if (settings.highContrast) {
      root.classList.add('high-contrast');
    } else {
      root.classList.remove('high-contrast');
    }
    
    // Larger text
    if (settings.largerText) {
      root.classList.add('large-text');
    } else {
      root.classList.remove('large-text');
    }
    
    // Keyboard navigation indicators
    if (settings.keyboardNav) {
      root.classList.add('show-focus');
    } else {
      root.classList.remove('show-focus');
    }
    
    saveSettings();
  }
  
  function saveSettings() {
    localStorage.setItem('crossPayAccessibility', JSON.stringify(settings));
  }
  
  function toggleSetting(key: keyof typeof settings) {
    settings[key] = !settings[key];
    applySettings();
  }
  
  function announceToScreenReader(message: string) {
    const announcer = document.getElementById('accessibility-announcer');
    if (announcer) {
      announcer.textContent = message;
      setTimeout(() => {
        announcer.textContent = '';
      }, 1000);
    }
  }
</script>

<!-- Accessibility Controls Button -->
<button
  class="accessibility-toggle cyber-btn-secondary"
  on:click={() => showControls = true}
  aria-label="Open accessibility controls"
  title="Accessibility Controls (Alt+A)"
>
  <span class="terminal-text">♿</span>
</button>

<!-- Accessibility Controls Modal -->
<CyberModal bind:open={showControls} title="ACCESSIBILITY_CONTROLS">
  <div class="space-y-6">
    <!-- Visual Settings -->
    <div class="cyber-card bg-cyber-surface-2 p-4">
      <div class="terminal-text text-cyber-mint text-sm mb-4">
        [VISUAL_SETTINGS]
      </div>
      
      <div class="space-y-3">
        <label class="flex items-center justify-between">
          <span class="terminal-text text-cyber-text-primary text-sm">
            HIGH_CONTRAST_MODE
          </span>
          <button
            class="accessibility-toggle-btn {settings.highContrast ? 'active' : ''}"
            on:click={() => {
              toggleSetting('highContrast');
              announceToScreenReader('High contrast mode ' + (settings.highContrast ? 'enabled' : 'disabled'));
            }}
            aria-pressed={settings.highContrast}
          >
            <div class="toggle-slider"></div>
          </button>
        </label>
        
        <label class="flex items-center justify-between">
          <span class="terminal-text text-cyber-text-primary text-sm">
            LARGER_TEXT_SIZE
          </span>
          <button
            class="accessibility-toggle-btn {settings.largerText ? 'active' : ''}"
            on:click={() => {
              toggleSetting('largerText');
              announceToScreenReader('Large text ' + (settings.largerText ? 'enabled' : 'disabled'));
            }}
            aria-pressed={settings.largerText}
          >
            <div class="toggle-slider"></div>
          </button>
        </label>
      </div>
    </div>
    
    <!-- Motion Settings -->
    <div class="cyber-card bg-cyber-surface-2 p-4">
      <div class="terminal-text text-cyber-mint text-sm mb-4">
        [MOTION_SETTINGS]
      </div>
      
      <div class="space-y-3">
        <label class="flex items-center justify-between">
          <span class="terminal-text text-cyber-text-primary text-sm">
            REDUCE_ANIMATIONS
          </span>
          <button
            class="accessibility-toggle-btn {settings.reducedMotion ? 'active' : ''}"
            on:click={() => {
              toggleSetting('reducedMotion');
              announceToScreenReader('Reduced motion ' + (settings.reducedMotion ? 'enabled' : 'disabled'));
            }}
            aria-pressed={settings.reducedMotion}
          >
            <div class="toggle-slider"></div>
          </button>
        </label>
      </div>
      
      <div class="terminal-text text-cyber-text-tertiary text-xs mt-2">
        > Disables scan lines, glitch effects, and animations
      </div>
    </div>
    
    <!-- Navigation Settings -->
    <div class="cyber-card bg-cyber-surface-2 p-4">
      <div class="terminal-text text-cyber-mint text-sm mb-4">
        [NAVIGATION_SETTINGS]
      </div>
      
      <div class="space-y-3">
        <label class="flex items-center justify-between">
          <span class="terminal-text text-cyber-text-primary text-sm">
            KEYBOARD_FOCUS_INDICATORS
          </span>
          <button
            class="accessibility-toggle-btn {settings.keyboardNav ? 'active' : ''}"
            on:click={() => {
              toggleSetting('keyboardNav');
              announceToScreenReader('Keyboard navigation indicators ' + (settings.keyboardNav ? 'enabled' : 'disabled'));
            }}
            aria-pressed={settings.keyboardNav}
          >
            <div class="toggle-slider"></div>
          </button>
        </label>
      </div>
      
      <div class="terminal-text text-cyber-text-tertiary text-xs mt-2">
        > Shows enhanced focus outlines for keyboard navigation
      </div>
    </div>
    
    <!-- Keyboard Shortcuts -->
    <div class="cyber-card bg-cyber-surface-2 p-4">
      <div class="terminal-text text-cyber-mint text-sm mb-4">
        [KEYBOARD_SHORTCUTS]
      </div>
      
      <div class="space-y-2 text-xs">
        <div class="flex justify-between">
          <span class="text-cyber-text-secondary">Command Palette:</span>
          <span class="terminal-text text-cyber-mint">⌘K</span>
        </div>
        <div class="flex justify-between">
          <span class="text-cyber-text-secondary">Home:</span>
          <span class="terminal-text text-cyber-mint">⌘⇧H</span>
        </div>
        <div class="flex justify-between">
          <span class="text-cyber-text-secondary">Payment:</span>
          <span class="terminal-text text-cyber-mint">⌘⇧P</span>
        </div>
        <div class="flex justify-between">
          <span class="text-cyber-text-secondary">Receipts:</span>
          <span class="terminal-text text-cyber-mint">⌘⇧R</span>
        </div>
        <div class="flex justify-between">
          <span class="text-cyber-text-secondary">Security:</span>
          <span class="terminal-text text-cyber-mint">⌘⇧S</span>
        </div>
        <div class="flex justify-between">
          <span class="text-cyber-text-secondary">AI Copilot:</span>
          <span class="terminal-text text-cyber-mint">⌘⇧A</span>
        </div>
      </div>
    </div>
  </div>
  
  <div slot="footer">
    <CyberButton variant="primary" on:click={() => showControls = false}>
      [SAVE_SETTINGS]
    </CyberButton>
  </div>
</CyberModal>

<!-- Screen Reader Announcer -->
<div id="accessibility-announcer" class="sr-only" aria-live="polite"></div>

<style>
  .accessibility-toggle {
    position: fixed;
    bottom: 1rem;
    right: 1rem;
    z-index: 900;
    padding: 0.75rem;
    min-height: auto;
    border-radius: 50%;
    aspect-ratio: 1;
  }
  
  .accessibility-toggle-btn {
    position: relative;
    width: 48px;
    height: 24px;
    border: 1px solid var(--cyber-border-mint);
    background: var(--cyber-surface-2);
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .accessibility-toggle-btn.active {
    background: var(--cyber-mint);
    border-color: var(--cyber-mint);
  }
  
  .toggle-slider {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 16px;
    height: 16px;
    background: var(--cyber-text-primary);
    transition: transform 0.2s;
  }
  
  .accessibility-toggle-btn.active .toggle-slider {
    transform: translateX(20px);
    background: var(--cyber-bg-primary);
  }
  
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  
  /* Global accessibility styles */
  :global(.reduce-motion) {
    --cyber-scan-line-duration: 0s;
    --cyber-glitch-duration: 0s;
    --cyber-typewriter-duration: 0s;
  }
  
  :global(.reduce-motion *) {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
  
  :global(.high-contrast) {
    --cyber-mint: rgba(159, 239, 223, 0.8);
    --cyber-lavender: rgba(201, 179, 255, 0.7);
    --cyber-text-primary: #FFFFFF;
    --cyber-text-secondary: #CCCCCC;
    --cyber-border-mint: rgba(159, 239, 223, 0.6);
    --cyber-border-lavender: rgba(201, 179, 255, 0.5);
  }
  
  :global(.large-text) {
    font-size: 120% !important;
  }
  
  :global(.large-text) .terminal-text {
    font-size: 110% !important;
  }
  
  :global(.show-focus *:focus) {
    outline: 2px solid var(--cyber-mint) !important;
    outline-offset: 2px !important;
  }
  
  @media (max-width: 480px) {
    .accessibility-toggle {
      bottom: 0.5rem;
      right: 0.5rem;
      padding: 0.5rem;
    }
  }
</style>