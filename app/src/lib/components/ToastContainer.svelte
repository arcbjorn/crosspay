<script lang="ts">
  import { toasts, removeToast } from '$lib/stores/toast';
  
  $: currentToasts = $toasts;
  
  const getToastClass = (type: string) => {
    switch (type) {
      case 'success': return 'alert-success';
      case 'error': return 'alert-error';
      case 'warning': return 'alert-warning';
      case 'info': return 'alert-info';
      default: return 'alert-info';
    }
  };

  const getToastIcon = (type: string) => {
    switch (type) {
      case 'success': return '✅';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      case 'info': return 'ℹ️';
      default: return 'ℹ️';
    }
  };
</script>

<div class="toast toast-end z-50">
  {#each currentToasts as toast (toast.id)}
    <div class="alert {getToastClass(toast.type)} shadow-lg min-w-80">
      <div class="flex items-center justify-between w-full">
        <div class="flex items-center gap-2">
          <span class="text-lg">{getToastIcon(toast.type)}</span>
          <span>{toast.message}</span>
        </div>
        {#if toast.dismissible}
          <button 
            class="btn btn-ghost btn-sm btn-circle"
            on:click={() => removeToast(toast.id)}
          >
            ✕
          </button>
        {/if}
      </div>
    </div>
  {/each}
</div>

<style>
  .toast {
    position: fixed;
    top: 1rem;
    right: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 400px;
  }
  
  .alert {
    animation: slideIn 0.3s ease-out;
  }
  
  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }
</style>