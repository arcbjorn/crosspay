<script lang="ts">
  import { onMount } from 'svelte';
  import CyberToast from './CyberToast.svelte';

  interface Toast {
    id: string;
    type: 'info' | 'success' | 'warning' | 'error';
    title?: string;
    message: string;
    duration?: number;
    dismissible?: boolean;
    showProgress?: boolean;
  }

  export let position: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' = 'top-right';
  export let maxToasts = 5;

  let toasts: Toast[] = [];

  $: positionClasses = {
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4'
  };

  function addToast(toast: Omit<Toast, 'id'>) {
    const id = Math.random().toString(36).substr(2, 9);
    const newToast: Toast = { id, ...toast };
    
    toasts = [newToast, ...toasts.slice(0, maxToasts - 1)];
    
    return id;
  }

  function removeToast(id: string) {
    toasts = toasts.filter(t => t.id !== id);
  }

  function clearAll() {
    toasts = [];
  }

  // Expose methods globally
  onMount(() => {
    if (typeof window !== 'undefined') {
      window.toast = {
        info: (message: string, title?: string, options?: Partial<Toast>) => 
          addToast({ type: 'info', message, title, ...options }),
        success: (message: string, title?: string, options?: Partial<Toast>) => 
          addToast({ type: 'success', message, title, ...options }),
        warning: (message: string, title?: string, options?: Partial<Toast>) => 
          addToast({ type: 'warning', message, title, ...options }),
        error: (message: string, title?: string, options?: Partial<Toast>) => 
          addToast({ type: 'error', message, title, ...options }),
        dismiss: removeToast,
        clear: clearAll
      };
    }
  });
</script>

<!-- Toast container -->
<div 
  class="fixed z-50 flex flex-col gap-2 pointer-events-none {positionClasses[position]}"
  class:items-end={position.includes('right')}
  class:items-start={position.includes('left')}
>
  {#each toasts as toast (toast.id)}
    <div class="pointer-events-auto">
      <CyberToast
        type={toast.type}
        title={toast.title || ''}
        message={toast.message}
        duration={toast.duration}
        dismissible={toast.dismissible}
        showProgress={toast.showProgress}
        on:dismiss={() => removeToast(toast.id)}
      />
    </div>
  {/each}
</div>

<!-- Terminal notification sound (optional) -->
<audio id="toast-sound" preload="none">
  <source src="data:audio/wav;base64,UklGRuACAABXQVZFZm10IBAAAAABAAEAIlYAAESsAAACABAAZGF0YeACAAABAAAA" type="audio/wav">
</audio>