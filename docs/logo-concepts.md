# Memzent Logo Concepts

> SVG logos ready to use. Open in browser, Figma, or any SVG editor. Export as PNG at any resolution.

---

## Concept 1: Neural Shield (Current Brand Direction)
*A shield with a neural/brain circuit pattern — represents memory + security.*

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120" width="120" height="120">
  <defs>
    <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#00f3ff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:1" />
    </linearGradient>
    <filter id="glow1">
      <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
      <feMerge><feMergeNode in="coloredBlur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
  </defs>
  <!-- Shield shape -->
  <path d="M60 10 L100 30 L100 65 C100 85 80 100 60 110 C40 100 20 85 20 65 L20 30 Z" 
        fill="url(#grad1)" opacity="0.15" stroke="url(#grad1)" stroke-width="2"/>
  <!-- Neural nodes -->
  <circle cx="60" cy="45" r="6" fill="#00f3ff" filter="url(#glow1)"/>
  <circle cx="42" cy="62" r="4" fill="#7c3aed" filter="url(#glow1)"/>
  <circle cx="78" cy="62" r="4" fill="#7c3aed" filter="url(#glow1)"/>
  <circle cx="48" cy="80" r="3.5" fill="#00f3ff" filter="url(#glow1)"/>
  <circle cx="72" cy="80" r="3.5" fill="#00f3ff" filter="url(#glow1)"/>
  <circle cx="60" cy="95" r="3" fill="#7c3aed" filter="url(#glow1)"/>
  <!-- Neural connections -->
  <line x1="60" y1="45" x2="42" y2="62" stroke="#00f3ff" stroke-width="1.5" opacity="0.6"/>
  <line x1="60" y1="45" x2="78" y2="62" stroke="#00f3ff" stroke-width="1.5" opacity="0.6"/>
  <line x1="42" y1="62" x2="48" y2="80" stroke="#7c3aed" stroke-width="1.5" opacity="0.6"/>
  <line x1="78" y1="62" x2="72" y2="80" stroke="#7c3aed" stroke-width="1.5" opacity="0.6"/>
  <line x1="42" y1="62" x2="72" y2="80" stroke="#00f3ff" stroke-width="1" opacity="0.3"/>
  <line x1="78" y1="62" x2="48" y2="80" stroke="#00f3ff" stroke-width="1" opacity="0.3"/>
  <line x1="48" y1="80" x2="60" y2="95" stroke="#7c3aed" stroke-width="1.5" opacity="0.6"/>
  <line x1="72" y1="80" x2="60" y2="95" stroke="#7c3aed" stroke-width="1.5" opacity="0.6"/>
</svg>
```

---

## Concept 2: Layered Cache "M"
*The letter M formed by stacked cache layers — represents the 4-layer architecture.*

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120" width="120" height="120">
  <defs>
    <linearGradient id="grad2a" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:#00f3ff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#00f3ff;stop-opacity:0.3" />
    </linearGradient>
    <linearGradient id="grad2b" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" style="stop-color:#7c3aed;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:0.3" />
    </linearGradient>
  </defs>
  <!-- Background rounded square -->
  <rect x="10" y="10" width="100" height="100" rx="24" fill="#0a0a1a" stroke="#1a1a3a" stroke-width="1"/>
  <!-- M shape from layers -->
  <path d="M30 90 L30 40 L45 60 L60 35 L75 60 L90 40 L90 90" 
        fill="none" stroke="url(#grad2a)" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/>
  <!-- Layer lines (representing cache layers) -->
  <line x1="25" y1="50" x2="95" y2="50" stroke="#00f3ff" stroke-width="0.5" opacity="0.2"/>
  <line x1="25" y1="60" x2="95" y2="60" stroke="#7c3aed" stroke-width="0.5" opacity="0.2"/>
  <line x1="25" y1="70" x2="95" y2="70" stroke="#00f3ff" stroke-width="0.5" opacity="0.2"/>
  <line x1="25" y1="80" x2="95" y2="80" stroke="#7c3aed" stroke-width="0.5" opacity="0.2"/>
  <!-- Glow dot at the peak -->
  <circle cx="60" cy="35" r="4" fill="#00f3ff" opacity="0.8"/>
</svg>
```

---

## Concept 3: Semantic Proxy Node
*A central node with orbiting connections — represents the proxy intercepting traffic.*

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120" width="120" height="120">
  <defs>
    <linearGradient id="grad3" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#00f3ff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:1" />
    </linearGradient>
    <filter id="glow3">
      <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
      <feMerge><feMergeNode in="coloredBlur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
  </defs>
  <!-- Background -->
  <rect x="10" y="10" width="100" height="100" rx="24" fill="#0a0a1a"/>
  <!-- Orbital rings -->
  <ellipse cx="60" cy="60" rx="35" ry="18" fill="none" stroke="#00f3ff" stroke-width="1" opacity="0.3" transform="rotate(-30 60 60)"/>
  <ellipse cx="60" cy="60" rx="35" ry="18" fill="none" stroke="#7c3aed" stroke-width="1" opacity="0.3" transform="rotate(30 60 60)"/>
  <ellipse cx="60" cy="60" rx="35" ry="18" fill="none" stroke="#00f3ff" stroke-width="1" opacity="0.3" transform="rotate(90 60 60)"/>
  <!-- Central core -->
  <circle cx="60" cy="60" r="14" fill="url(#grad3)" filter="url(#glow3)"/>
  <!-- M letter in center -->
  <text x="60" y="66" text-anchor="middle" fill="#0a0a1a" font-family="system-ui" font-size="16" font-weight="900">M</text>
  <!-- Orbiting dots (clients/providers) -->
  <circle cx="32" cy="42" r="4" fill="#00f3ff" opacity="0.8"/>
  <circle cx="88" cy="42" r="4" fill="#7c3aed" opacity="0.8"/>
  <circle cx="60" cy="25" r="4" fill="#00f3ff" opacity="0.8"/>
  <circle cx="32" cy="78" r="3" fill="#7c3aed" opacity="0.6"/>
  <circle cx="88" cy="78" r="3" fill="#00f3ff" opacity="0.6"/>
  <circle cx="60" cy="95" r="3" fill="#7c3aed" opacity="0.6"/>
</svg>
```

---

## Concept 4: Memory Hexagon
*Hexagonal brain cell shape — represents memory, intelligence, and modularity.*

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120" width="120" height="120">
  <defs>
    <linearGradient id="grad4" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#00f3ff;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#7c3aed;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#00f3ff;stop-opacity:1" />
    </linearGradient>
    <filter id="glow4">
      <feGaussianBlur stdDeviation="2.5" result="coloredBlur"/>
      <feMerge><feMergeNode in="coloredBlur"/><feMergeNode in="SourceGraphic"/></feMerge>
    </filter>
  </defs>
  <!-- Outer hexagon -->
  <polygon points="60,12 100,35 100,85 60,108 20,85 20,35" 
           fill="#0a0a1a" stroke="url(#grad4)" stroke-width="2.5" filter="url(#glow4)"/>
  <!-- Inner hexagon -->
  <polygon points="60,30 82,43 82,77 60,90 38,77 38,43" 
           fill="none" stroke="#00f3ff" stroke-width="1" opacity="0.3"/>
  <!-- Brain/circuit paths inside -->
  <path d="M45 50 Q60 40 75 50 Q80 60 75 70 Q60 80 45 70 Q40 60 45 50" 
        fill="none" stroke="#00f3ff" stroke-width="1.5" opacity="0.6"/>
  <path d="M52 55 Q60 48 68 55" fill="none" stroke="#7c3aed" stroke-width="1.5" opacity="0.8"/>
  <path d="M52 65 Q60 72 68 65" fill="none" stroke="#7c3aed" stroke-width="1.5" opacity="0.8"/>
  <!-- Central dot -->
  <circle cx="60" cy="60" r="5" fill="#00f3ff" filter="url(#glow4)"/>
  <!-- Corner accents -->
  <circle cx="60" cy="12" r="2" fill="#00f3ff" opacity="0.6"/>
  <circle cx="100" cy="35" r="2" fill="#7c3aed" opacity="0.6"/>
  <circle cx="100" cy="85" r="2" fill="#00f3ff" opacity="0.6"/>
  <circle cx="60" cy="108" r="2" fill="#7c3aed" opacity="0.6"/>
  <circle cx="20" cy="85" r="2" fill="#00f3ff" opacity="0.6"/>
  <circle cx="20" cy="35" r="2" fill="#7c3aed" opacity="0.6"/>
</svg>
```

---

## Concept 5: Speed Bolt + Brain (Minimal)
*Lightning bolt merged with a brain outline — speed + intelligence. Great for favicons.*

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120" width="120" height="120">
  <defs>
    <linearGradient id="grad5" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#00f3ff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:1" />
    </linearGradient>
  </defs>
  <!-- Background circle -->
  <circle cx="60" cy="60" r="50" fill="#0a0a1a" stroke="url(#grad5)" stroke-width="2"/>
  <!-- Stylized bolt/M hybrid -->
  <path d="M42 30 L62 30 L52 55 L68 55 L45 95 L55 65 L38 65 Z" 
        fill="url(#grad5)"/>
  <!-- Right side brain curve -->
  <path d="M65 35 C85 35 90 50 85 60 C90 70 85 85 65 90" 
        fill="none" stroke="#00f3ff" stroke-width="2" opacity="0.5" stroke-linecap="round"/>
  <!-- Brain wrinkle -->
  <path d="M70 50 C78 55 78 65 70 70" 
        fill="none" stroke="#7c3aed" stroke-width="1.5" opacity="0.4" stroke-linecap="round"/>
</svg>
```

---

## Color Palette

| Color | Hex | Usage |
|-------|-----|-------|
| Memzent Glow (Cyan) | `#00f3ff` | Primary brand, highlights, CTAs |
| Memzent Purple | `#7c3aed` | Secondary, gradients, accents |
| Memzent Accent (Teal) | `#0ea5e9` | Tertiary highlights |
| Dark Background | `#0a0a1a` | Logo backgrounds |
| Light variant | `#ffffff` | Logo on dark backgrounds |

---

## Wordmark Variants

### Dark background (primary)
```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 60" width="300" height="60">
  <defs>
    <linearGradient id="wg1" x1="0%" y1="0%" x2="100%" y2="0%">
      <stop offset="0%" style="stop-color:#00f3ff" />
      <stop offset="100%" style="stop-color:#7c3aed" />
    </linearGradient>
  </defs>
  <!-- Icon -->
  <rect x="5" y="10" width="40" height="40" rx="10" fill="url(#wg1)"/>
  <text x="25" y="37" text-anchor="middle" fill="#0a0a1a" font-family="system-ui" font-size="20" font-weight="900">M</text>
  <!-- Wordmark -->
  <text x="55" y="40" fill="#ffffff" font-family="system-ui" font-size="28" font-weight="900" letter-spacing="-1">MEMZENT</text>
  <!-- Tagline -->
  <text x="55" y="55" fill="#ffffff" font-family="system-ui" font-size="9" font-weight="600" opacity="0.5" letter-spacing="2">.AI</text>
</svg>
```

---

## Recommendations

| Use Case | Best Concept |
|----------|-------------|
| **Favicon (16×16, 32×32)** | #5 Speed Bolt (simple at small sizes) |
| **App icon / Dashboard** | #1 Neural Shield (current direction) |
| **LinkedIn profile** | #3 Semantic Proxy Node (eye-catching) |
| **GitHub repo** | #2 Layered Cache M (developer-friendly) |
| **Pitch deck / Enterprise** | #4 Memory Hexagon (professional) |
| **All with text** | Wordmark variant below any icon |

### Export Tips
- Save each SVG block as a `.svg` file
- Open in browser → right-click → "Save as image" for quick PNG
- For hi-res: use Figma (paste SVG) or Inkscape to export at 512×512, 1024×1024
- LinkedIn profile photo: 400×400px minimum
