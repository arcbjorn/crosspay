# CrossPay Mini App

> Ultra-lightweight Base MiniKit application for viral payments

A single-file HTML app optimized for the Base MiniKit platform, featuring viral payment mechanics, QR code generation, and gamified user engagement in under 100KB.

## 🎯 Overview

The CrossPay Mini App is a standalone application designed specifically for the Base MiniKit ecosystem. It focuses on quick USDC payments with viral features to drive user adoption and engagement.

### Key Features
- **Single File**: 98KB HTML file with inline CSS and JavaScript
- **Viral Mechanics**: Payment streaks, referral codes, and achievements
- **QR Payments**: Generate and share payment request QR codes
- **Local Storage**: Persistent user data and preferences
- **Mobile Optimized**: Touch-first design for mobile wallets

## 📦 File Structure

```
mini/
├── index.html        # Main app file (98KB)
├── package.json      # Minimal package metadata
└── README.md         # This documentation
```

## 🚀 Quick Start

### Local Development
```bash
# Serve the app locally
cd mini
python3 -m http.server 8080

# Open in browser
open http://localhost:8080
```

### Size Check
```bash
# Verify file size (must be <100KB for MiniKit)
du -h index.html
# Output: 98K index.html
```

### Base MiniKit Integration
The app is designed to work within the Base wallet MiniKit environment:
- Loads as a single HTML file
- Optimized for mobile viewing
- Integrates with Base wallet APIs

## 🎮 Viral Features

### Payment Streaks
```javascript
// Track consecutive payment days
let paymentStreak = parseInt(localStorage.getItem('paymentStreak') || '7');
```

- **Streak Counter**: Days of consecutive payments
- **Visual Progress**: Animated streak display
- **Achievement Badges**: Milestone rewards (7, 30, 100 days)
- **Leaderboard**: Top 10% achievement status

### Referral System
```javascript
// Generate unique referral codes
function generateReferralCode() {
  const code = 'CROSS' + Math.random().toString(36).substr(2, 6).toUpperCase();
  localStorage.setItem('referralCode', code);
  return code;
}
```

- **Unique Codes**: CROSS + 6 character suffix
- **Share Links**: Deep linking with referral tracking
- **Bonus Rewards**: Incentives for successful referrals
- **Social Sharing**: Native share API integration

### Achievement System
- **Payment Milestones**: 10, 50, 100, 500 payments
- **Streak Rewards**: Weekly, monthly streak bonuses
- **Early Adopter**: Special badges for first users
- **VIP Status**: Exclusive benefits for top users

## 💰 Payment Features

### Quick Pay Interface
```javascript
// Simplified payment creation
const payment = {
  recipient: '0x...',
  amount: '10.00',
  token: 'USDC',
  chain: 'base-sepolia'
};
```

- **One-Tap Payments**: Minimal friction payment flow
- **USDC Focus**: Stablecoin-first payment experience
- **Address Book**: Save frequent recipients
- **Amount Presets**: Quick selection ($5, $10, $25, $50)

### QR Code Generation
```javascript
function generateQR() {
  const qrData = {
    type: 'payment_request',
    recipient: userAddress,
    amount: amount,
    referralCode: referralCode
  };
  
  // Display QR code for scanning
  displayQRCode(JSON.stringify(qrData));
}
```

- **Payment Requests**: Generate QR codes for receiving payments
- **Deep Links**: App-specific URLs for mobile sharing
- **Social Integration**: Share to messaging apps
- **Offline Compatible**: QR codes work without internet

## 🎨 UI/UX Design

### Design Constraints
- **100KB Limit**: All assets must fit in single HTML file
- **Mobile First**: Optimized for smartphone screens
- **Touch Friendly**: Large tap targets and gestures
- **Fast Loading**: Minimal JavaScript and CSS

### Color Scheme
```css
:root {
  --primary: #667eea;
  --secondary: #764ba2;
  --success: #00d4aa;
  --warning: #f59e0b;
  --background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
```

### Typography
- **System Fonts**: Platform-specific font stack
- **Icon Fonts**: Unicode symbols for minimal size
- **Responsive Text**: Scales with device size

## 📱 Mobile Optimization

### Touch Interactions
```css
.btn {
  min-height: 44px;  /* iOS minimum touch target */
  padding: 12px 24px;
  border-radius: 8px;
}
```

- **Large Buttons**: 44px minimum height for accessibility
- **Touch Feedback**: Visual response to taps
- **Swipe Gestures**: Navigate between screens
- **Haptic Feedback**: Vibration on successful actions

### Performance
- **Lazy Loading**: Images and content loaded on demand
- **Local Storage**: Cache user data and preferences
- **Minimal JavaScript**: Vanilla JS for maximum performance
- **CSS Animations**: Hardware-accelerated transitions

## 🔧 Technical Implementation

### Single File Architecture
```html
<!DOCTYPE html>
<html>
<head>
  <style>/* All CSS inline */</style>
</head>
<body>
  <div id="app"><!-- All HTML structure --></div>
  <script>/* All JavaScript inline */</script>
</body>
</html>
```

### State Management
```javascript
// Simple state object
const appState = {
  user: {
    paymentStreak: 7,
    totalPayments: 23,
    referralCode: 'CROSS123ABC',
    achievements: ['7day', 'first_payment', 'top_10']
  },
  payments: [],
  settings: {
    theme: 'dark',
    notifications: true
  }
};
```

### Local Storage
```javascript
// Persist user data
function saveState() {
  localStorage.setItem('crossPayState', JSON.stringify(appState));
}

function loadState() {
  const saved = localStorage.getItem('crossPayState');
  if (saved) {
    Object.assign(appState, JSON.parse(saved));
  }
}
```

## 🔗 Base MiniKit Integration

### Wallet Connection
```javascript
// Connect to Base wallet
async function connectBaseWallet() {
  if (window.ethereum?.isBase) {
    const accounts = await window.ethereum.request({
      method: 'eth_requestAccounts'
    });
    return accounts[0];
  }
  throw new Error('Base wallet not detected');
}
```

### Chain Configuration
```javascript
const BASE_SEPOLIA = {
  chainId: '0x14a34',  // 84532
  chainName: 'Base Sepolia',
  rpcUrls: ['https://sepolia.base.org'],
  blockExplorerUrls: ['https://sepolia-explorer.base.org']
};
```

## 📊 Analytics & Metrics

### User Engagement
- **Session Duration**: Time spent in app
- **Payment Frequency**: Payments per day/week
- **Streak Retention**: Users maintaining streaks
- **Referral Success**: Successful referral conversions

### Viral Metrics
- **Share Rate**: Percentage of users sharing
- **Referral Click Rate**: Referral link engagement
- **Achievement Completion**: Badge unlock rates
- **Return Rate**: Daily/weekly active users

## 🚀 Deployment

### MiniKit Submission
1. **Size Verification**: Ensure <100KB total size
2. **Testing**: Validate on multiple devices
3. **Base Review**: Submit to Base MiniKit store
4. **Distribution**: Available in Base wallet

### Update Strategy
```javascript
// Version checking
const APP_VERSION = '1.0.0';
const LATEST_VERSION_URL = 'https://api.crosspay.app/mini/version';

async function checkForUpdates() {
  // Prompt user for manual update
  const latest = await fetch(LATEST_VERSION_URL);
  if (latest.version > APP_VERSION) {
    showUpdatePrompt();
  }
}
```

## 🔐 Security

### Data Protection
- **Local Storage Only**: No server-side data storage
- **No Private Keys**: Wallet handles key management
- **Input Validation**: Sanitize all user inputs
- **XSS Prevention**: No dynamic HTML generation

### Privacy
- **No Tracking**: No third-party analytics
- **Opt-in Sharing**: User controls data sharing
- **Anonymous IDs**: No personally identifiable information

## 📈 Future Enhancements

### v2.0 Features
- **Camera Integration**: QR code scanning
- **Push Notifications**: Payment alerts
- **Social Login**: Connect social accounts
- **Multi-language**: Support for major languages

### Advanced Viral Features
- **Team Challenges**: Group payment goals
- **Seasonal Events**: Limited-time achievements
- **NFT Rewards**: Unique collectibles for milestones
- **Integration APIs**: Connect with other Base apps

## 🐛 Troubleshooting

### Common Issues

**App won't load in Base wallet**
```javascript
// Check MiniKit compatibility
if (!window.miniKit) {
  console.error('MiniKit not available');
}
```

**QR codes not generating**
```javascript
// Fallback to text-based sharing
if (!qrCodeSupported) {
  showTextShareLink(paymentData);
}
```

**Local storage full**
```javascript
// Clear old data
function cleanupStorage() {
  const oldData = JSON.parse(localStorage.getItem('crossPayState'));
  // Keep only recent payments
  oldData.payments = oldData.payments.slice(-50);
  localStorage.setItem('crossPayState', JSON.stringify(oldData));
}
```

## 📄 License

MIT License - optimized for viral growth and community adoption.

---

**Ultra Lightweight** ⚡ | **Viral Ready** 🚀 | **Mobile First** 📱