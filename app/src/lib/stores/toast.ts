import { writable } from 'svelte/store';

export interface Toast {
  id: string;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
  duration?: number;
  dismissible?: boolean;
}

export const toasts = writable<Toast[]>([]);

export function addToast(
  message: string, 
  type: Toast['type'] = 'info', 
  duration: number = 5000,
  dismissible: boolean = true
): string {
  const id = Math.random().toString(36).substr(2, 9);
  const toast: Toast = { id, message, type, duration, dismissible };
  
  toasts.update(items => [...items, toast]);
  
  // Auto-remove after duration
  if (duration > 0) {
    setTimeout(() => {
      removeToast(id);
    }, duration);
  }
  
  return id;
}

export function removeToast(id: string): void {
  toasts.update(items => items.filter(item => item.id !== id));
}

export function clearToasts(): void {
  toasts.set([]);
}

// Convenience functions
export const successToast = (message: string, duration?: number) => 
  addToast(message, 'success', duration);

export const errorToast = (message: string, duration?: number) => 
  addToast(message, 'error', duration);

export const warningToast = (message: string, duration?: number) => 
  addToast(message, 'warning', duration);

export const infoToast = (message: string, duration?: number) => 
  addToast(message, 'info', duration);