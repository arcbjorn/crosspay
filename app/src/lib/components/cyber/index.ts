// Core Components
export { default as CyberButton } from './CyberButton.svelte';
export { default as CyberInput } from './CyberInput.svelte';
export { default as CyberCard } from './CyberCard.svelte';
export { default as CyberModal } from './CyberModal.svelte';
export { default as CyberTable } from './CyberTable.svelte';
export { default as CyberLoader } from './CyberLoader.svelte';
export { default as CyberErrorBoundary } from './CyberErrorBoundary.svelte';

// New Components
export { default as CyberChart } from './CyberChart.svelte';
export { default as CyberToast } from './CyberToast.svelte';
export { default as ToastContainer } from './ToastContainer.svelte';
export { default as DataStream } from './DataStream.svelte';
export { default as GridSystem } from './GridSystem.svelte';
export { default as NetworkGrid } from './NetworkGrid.svelte';
export { default as Terminal } from './Terminal.svelte';

// Composite Components
export { default as PaymentTerminal } from './PaymentTerminal.svelte';
export { default as SecurityMatrix } from './SecurityMatrix.svelte';
export { default as RiskMeter } from './RiskMeter.svelte';
export { default as InvoiceParser } from './InvoiceParser.svelte';

// Component Types
export interface CyberButtonProps {
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  href?: string;
  glitch?: boolean;
}

export interface CyberInputProps {
  value?: string;
  type?: string;
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  label?: string;
  terminal?: boolean;
}

export interface CyberCardProps {
  variant?: 'default' | 'mint' | 'lavender' | 'danger';
  padding?: 'sm' | 'md' | 'lg';
  scan?: boolean;
  glow?: boolean;
  matrix?: boolean;
}

export interface PaymentTerminalProps {
  loading?: boolean;
  networks?: Array<{id: string, name: string, symbol: string}>;
}

export interface SecurityMatrixProps {
  validators?: Array<{
    id: string;
    address: string;
    status: 'active' | 'inactive' | 'syncing';
    stake: number;
    uptime: number;
  }>;
  consensusStatus?: 'finalized' | 'pending' | 'error';
  securityLevel?: 'low' | 'medium' | 'high' | 'maximum';
}

export interface RiskMeterProps {
  riskScore?: number;
  factors?: Array<{
    name: string;
    impact: number;
    description: string;
  }>;
  loading?: boolean;
}