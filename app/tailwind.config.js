/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	darkMode: 'class',
	theme: {
		extend: {
			colors: {
				// Cyberpunk color palette
				cyber: {
					// Background colors
					'bg-primary': '#0A0A0B',
					'bg-secondary': '#141416',
					'surface-1': '#1C1C1F',
					'surface-2': '#252528',

					// Accent colors (tame pastels with opacity)
					mint: 'rgba(159, 239, 223, 0.4)',
					'mint-hover': 'rgba(159, 239, 223, 0.6)',
					lavender: 'rgba(201, 179, 255, 0.35)',
					'lavender-hover': 'rgba(201, 179, 255, 0.5)',

					// State colors
					warning: 'rgba(255, 228, 181, 0.45)',
					error: 'rgba(255, 179, 186, 0.4)',
					success: 'rgba(179, 255, 186, 0.35)',

					// Text hierarchy
					'text-primary': '#E0E0E0',
					'text-secondary': '#A0A0A0',
					'text-tertiary': '#707070',

					// Border colors
					'border-mint': 'rgba(159, 239, 223, 0.3)',
					'border-lavender': 'rgba(201, 179, 255, 0.25)'
				}
			},
			fontFamily: {
				mono: ['JetBrains Mono', 'Space Mono', 'ui-monospace', 'monospace'],
				sans: ['Inter', 'IBM Plex Sans', 'ui-sans-serif', 'system-ui']
			},
			fontSize: {
				xs: ['12px', { lineHeight: '18px', letterSpacing: '0.02em' }],
				sm: ['14px', { lineHeight: '21px', letterSpacing: '0.015em' }],
				base: ['16px', { lineHeight: '24px', letterSpacing: '0.01em' }],
				lg: ['20px', { lineHeight: '24px', letterSpacing: '0.005em' }],
				xl: ['24px', { lineHeight: '28.8px', letterSpacing: '0em' }],
				'2xl': ['32px', { lineHeight: '38.4px', letterSpacing: '-0.005em' }]
			},
			animation: {
				'scan-line': 'scanLine 3s linear infinite',
				glitch: 'glitch 0.3s ease-in-out',
				typewriter: 'typewriter 2s steps(20) forwards',
				'cursor-blink': 'cursorBlink 1s step-end infinite',
				'matrix-rain': 'matrixRain 10s linear infinite',
				'terminal-glow': 'terminalGlow 2s ease-in-out infinite alternate'
			},
			keyframes: {
				scanLine: {
					'0%': { transform: 'translateY(-100%)' },
					'100%': { transform: 'translateY(100vh)' }
				},
				glitch: {
					'0%, 100%': { transform: 'translate(0)' },
					'20%': { transform: 'translate(-2px, 2px)' },
					'40%': { transform: 'translate(-2px, -2px)' },
					'60%': { transform: 'translate(2px, 2px)' },
					'80%': { transform: 'translate(2px, -2px)' }
				},
				typewriter: {
					'0%': { width: '0ch' },
					'100%': { width: '100%' }
				},
				cursorBlink: {
					'0%, 50%': { borderColor: 'transparent' },
					'51%, 100%': { borderColor: 'rgba(159, 239, 223, 0.4)' }
				},
				matrixRain: {
					'0%': { transform: 'translateY(-100%)' },
					'100%': { transform: 'translateY(100vh)' }
				},
				terminalGlow: {
					'0%': { boxShadow: '0 0 5px rgba(159, 239, 223, 0.3)' },
					'100%': { boxShadow: '0 0 20px rgba(159, 239, 223, 0.6)' }
				}
			},
			backgroundImage: {
				'dot-matrix': 'radial-gradient(rgba(159, 239, 223, 0.1) 1px, transparent 1px)',
				'scan-lines':
					'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(159, 239, 223, 0.03) 2px, rgba(159, 239, 223, 0.03) 4px)',
				noise: `url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noiseFilter'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noiseFilter)' opacity='0.02'/%3E%3C/svg%3E")`
			},
			backgroundSize: {
				'dot-matrix': '20px 20px'
			},
			spacing: {
				grid: '8px'
			},
			gridTemplateColumns: {
				16: 'repeat(16, minmax(0, 1fr))'
			}
		}
	},
	plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
};
