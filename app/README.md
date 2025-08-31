# CrossPay Frontend Application

> Modern SvelteKit web application for CrossPay Protocol

A responsive web interface for creating, managing, and viewing payments across multiple blockchain networks. Built with SvelteKit, TypeScript, and TailwindCSS.

## 🎯 Features

### Core Functionality
- **Wallet Integration**: MetaMask connection with automatic network switching
- **Payment Creation**: Intuitive form with ENS name resolution support
- **Receipt Management**: View payment history and transaction status
- **Multi-Network**: Support for Lisk Sepolia and Base Sepolia
- **Real-time Updates**: Live transaction status and balance updates

### User Experience
- **Responsive Design**: Mobile-first approach with TailwindCSS
- **Dark/Light Mode**: Theme switching with DaisyUI components
- **Toast Notifications**: User-friendly transaction feedback
- **Loading States**: Smooth loading animations and skeletons
- **Error Handling**: Comprehensive error messaging and recovery

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and pnpm 8+
- MetaMask or compatible wallet
- Testnet ETH for Lisk/Base Sepolia

### Development Setup
```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Open browser
open http://localhost:5173
```

### Build for Production
```bash
# Build the application
pnpm build

# Preview production build
pnpm preview

# Run type checking
pnpm typecheck
```

## 📦 Project Structure

```
src/
├── routes/                 # SvelteKit pages
│   ├── +layout.svelte     # Root layout with navigation
│   ├── +page.svelte       # Landing page
│   ├── pay/              
│   │   └── +page.svelte   # Payment creation form
│   └── receipts/
│       └── +page.svelte   # Payment history
├── lib/
│   ├── components/        # Reusable Svelte components
│   │   └── WalletConnect.svelte
│   └── stores/           # Global state management
│       ├── wallet.ts     # Wallet connection state
│       └── chain.ts      # Network configuration
├── app.css               # Global styles with Tailwind
└── app.html              # HTML template
```

## 🔧 Key Components

### WalletConnect Component
Handles wallet connection and network management:

```svelte
<script lang="ts">
  import { walletStore, connectWallet, disconnectWallet } from '$lib/stores/wallet';
  import { chainStore, switchChain } from '$lib/stores/chain';
  
  $: wallet = $walletStore;
</script>

<WalletConnect />
```

### Wallet Store
Manages wallet connection state:

```typescript
interface WalletState {
  isConnected: boolean;
  address?: Address;
  chainId?: number;
  balance?: bigint;
  isConnecting: boolean;
  error?: string;
}

export const walletStore = writable<WalletState>(initialState);
```

### Chain Configuration
Network-specific settings:

```typescript
export const SUPPORTED_CHAINS = {
  4202: {
    name: 'Lisk Sepolia',
    rpcUrl: 'https://rpc.sepolia-api.lisk.com',
    blockExplorer: 'https://sepolia-blockscout.lisk.com',
    faucet: 'https://sepolia-faucet.lisk.com'
  },
  84532: {
    name: 'Base Sepolia',
    rpcUrl: 'https://sepolia.base.org',
    blockExplorer: 'https://sepolia-explorer.base.org'
  }
};
```

## 🎨 UI/UX Design

### Design System
- **TailwindCSS**: Utility-first CSS framework
- **DaisyUI**: Pre-built component themes
- **Custom Components**: Consistent design tokens

### Color Palette
```css
:root {
  --primary: #3b82f6;
  --primary-focus: #2563eb;
  --secondary: #64748b;
  --accent: #06b6d4;
  --neutral: #1f2937;
}
```

### Typography
- **Headings**: Inter font family
- **Body**: System font stack
- **Code**: JetBrains Mono for addresses

## 📱 Pages Overview

### Landing Page (`/`)
- Hero section with CrossPay branding
- Feature highlights (Privacy, Cross-chain, Verifiable)
- Wallet connection prompt for new users
- Quick navigation to payment creation

### Payment Creation (`/pay`)
- Network selection dropdown (Lisk/Base Sepolia)
- Recipient address input with ENS resolution
- Token selection (ETH, mock USDC)
- Amount input with fee calculation
- Metadata URI field for attachments
- Real-time validation and error handling

### Payment History (`/receipts`)
- Table view of all user payments
- Filtering by status and network
- Individual receipt details
- Transaction links to block explorers
- Actions for pending payments (complete/refund)

## 🔌 Wallet Integration

### Supported Wallets
- **MetaMask**: Primary integration
- **WalletConnect**: Prepared for future implementation

### Network Management
```typescript
export const switchChain = async (chainId: number) => {
  try {
    await window.ethereum.request({
      method: 'wallet_switchEthereumChain',
      params: [{ chainId: `0x${chainId.toString(16)}` }],
    });
  } catch (error: any) {
    if (error.code === 4902) {
      throw new Error(`Chain ${chainId} not found in wallet`);
    }
    throw error;
  }
};
```

### Balance Tracking
```typescript
const updateBalance = async (address: string) => {
  const balance = await window.ethereum.request({
    method: 'eth_getBalance',
    params: [address, 'latest']
  });
  
  walletStore.update(state => ({
    ...state,
    balance: BigInt(balance)
  }));
};
```

## 🧪 Testing

### Unit Tests (Vitest)
```bash
# Run component tests
pnpm test

# Run tests in watch mode
pnpm test:watch

# Generate coverage report
pnpm test:coverage
```

### E2E Tests (Playwright)
```bash
# Install browsers
pnpm playwright install

# Run end-to-end tests
pnpm test:e2e

# Run tests in headed mode
pnpm test:e2e --headed
```

### Test Structure
```
src/lib/components/__tests__/
├── WalletConnect.test.ts      # Component unit tests
├── PaymentForm.test.ts        # Payment form validation
└── ReceiptView.test.ts        # Receipt display logic

tests/
├── payment-flow.spec.ts       # E2E payment creation
├── wallet-connection.spec.ts  # Wallet integration tests
└── navigation.spec.ts         # Page routing tests
```

## 🚀 Deployment

### Environment Variables
```bash
# Production environment (.env.production)
PUBLIC_PAYMENT_CORE_ADDRESS_4202=0x...     # Lisk Sepolia
PUBLIC_PAYMENT_CORE_ADDRESS_84532=0x...    # Base Sepolia
PUBLIC_WALLETCONNECT_PROJECT_ID=...        # WalletConnect
PUBLIC_ENABLE_ANALYTICS=true               # Analytics
```

### Build Optimization
```bash
# Build with SvelteKit adapter
pnpm build

# Analyze bundle size
pnpm run build:analyze

# Preview production build
pnpm preview
```

### Deployment Platforms
- **Vercel**: Recommended for SvelteKit (auto-deployment)
- **Netlify**: Static site hosting with forms
- **IPFS**: Decentralized hosting via Fleek

## 🔍 Performance

### Core Web Vitals
- **LCP**: < 2.5s (optimized images and fonts)
- **FID**: < 100ms (minimal JavaScript)
- **CLS**: < 0.1 (reserved space for dynamic content)

### Optimization Strategies
- **Code Splitting**: Route-based splitting
- **Image Optimization**: WebP format with fallbacks
- **Font Loading**: Preload critical fonts
- **Caching**: Service worker for static assets

## 🛠️ Development Tools

### Code Quality
```bash
# Linting with ESLint
pnpm lint

# Format with Prettier
pnpm format

# Type checking
pnpm typecheck

# Svelte check
pnpm check
```

### Dev Tools Configuration
```json
{
  "extends": "./.svelte-kit/tsconfig.json",
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true
  }
}
```

## 🔐 Security Considerations

### Input Validation
- **Address Validation**: Ethereum address format checking
- **Amount Validation**: Positive number validation with max decimals
- **XSS Prevention**: Automatic escaping in Svelte templates

### Wallet Security
- **Permission Requests**: Minimal permission scope
- **Transaction Signing**: Clear transaction details before signing
- **Network Verification**: Chain ID validation before transactions

## 🚀 Future Enhancements

### Module 2+ Features
- **ENS Integration**: Real ENS name resolution
- **Token Lists**: Dynamic token discovery
- **Transaction History**: On-chain event parsing
- **Push Notifications**: Payment status updates
- **Mobile App**: React Native or Flutter

### User Experience
- **Dark Mode**: Full theme customization
- **Internationalization**: Multi-language support
- **Accessibility**: WCAG 2.1 AA compliance
- **Progressive Web App**: Offline functionality

## 📄 License

MIT License - open source and community-driven.

---

**Modern Stack** ⚡ | **Type Safe** 🛡️ | **Mobile Ready** 📱