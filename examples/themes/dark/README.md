# Dark Theme

A lush, modern dark theme showcasing cutting-edge CSS features while maintaining exceptional readability. Built for the modern web with OKLCH colors, CSS cascade layers, container queries, and fluid typography.

## Features

### Modern CSS Technologies
- **OKLCH Color Space**: Perceptually uniform colors for consistent brightness and vibrant accents
- **Cascade Layers**: Organized CSS architecture with `@layer` for predictable specificity
- **Fluid Typography**: `clamp()` for responsive text that scales smoothly across all viewports
- **Container Queries**: Layout components that adapt to their container size
- **Logical Properties**: `inline`/`block` properties for better internationalization
- **Modern Selectors**: `:has()`, `:where()`, `:is()` for powerful styling patterns

### Design Features
- **Lush Color Palette**: Rich dark blues and purples with electric accents (cyan, violet, pink, emerald)
- **Glassmorphism**: Frosted glass effects with `backdrop-filter` on sidebar
- **Gradient Text**: Multi-color gradients on headings using `background-clip: text`
- **Smooth Animations**: Subtle fade-in animations with `prefers-reduced-motion` support
- **Modern Cards**: Elevated entry cards with hover effects and shadows
- **Optimal Readability**: High contrast ratios, generous line-height, careful color choices

### Accessibility
- **Color Scheme**: Properly declares `color-scheme: dark` for native controls
- **Focus Indicators**: Visible focus outlines on all interactive elements
- **Motion Preferences**: Respects `prefers-reduced-motion` for accessibility
- **Semantic HTML**: Proper heading hierarchy and landmark regions
- **Text Rendering**: Optimized font smoothing for dark backgrounds

## Theme Structure

```
dark/
├── template.html        # HTML template with semantic markup
├── static/
│   ├── style.css       # Modern CSS with layers and advanced features
│   └── feed-icon.svg   # RSS feed icon
└── README.md           # This file
```

## Usage

### Basic Setup

1. Copy the dark theme to your project:
```bash
mkdir -p themes
cp -r examples/themes/dark themes/
```

2. Update your `config.ini`:
```ini
[planet]
name = My Planet
link = https://planet.example.com
owner_name = Your Name
template = ./themes/dark/template.html
```

3. Generate your site:
```bash
rp generate
```

## Customization

### Color Palette

The theme uses OKLCH colors defined in CSS custom properties. Edit these in `static/style.css`:

```css
:root {
  /* Background layers */
  --bg-primary: oklch(0.18 0.02 270);       /* Deep indigo-black */
  --bg-secondary: oklch(0.22 0.025 270);    /* Slightly lighter */
  --bg-tertiary: oklch(0.28 0.03 270);      /* Surface level */

  /* Accent colors - vibrant and lush */
  --accent-cyan: oklch(0.75 0.15 195);      /* Electric cyan */
  --accent-violet: oklch(0.65 0.22 285);    /* Rich violet */
  --accent-pink: oklch(0.70 0.20 340);      /* Hot pink */
  --accent-emerald: oklch(0.70 0.15 155);   /* Lush emerald */
}
```

**OKLCH Format**: `oklch(lightness chroma hue / alpha)`
- **Lightness**: 0 (black) to 1 (white)
- **Chroma**: Color intensity (0 = gray, higher = more saturated)
- **Hue**: 0-360 degrees (0=red, 120=green, 240=blue, 285=violet, 195=cyan)

### Typography

Fluid typography automatically scales between viewports:

```css
:root {
  /* Format: clamp(min, preferred, max) */
  --text-base: clamp(1rem, 0.95rem + 0.25vw, 1.125rem);
  --text-xl: clamp(1.5rem, 1.35rem + 0.75vw, 1.875rem);
}
```

### Cascade Layers

The CSS is organized into layers for maintainability:

1. `@layer reset` - Base resets
2. `@layer base` - Body and document styles
3. `@layer typography` - Text and heading styles
4. `@layer layout` - Page structure
5. `@layer components` - UI components
6. `@layer responsive` - Media queries
7. `@layer animations` - Motion and transitions

To override styles, add your own layer:

```css
@layer overrides {
  .entry {
    background: oklch(0.25 0.03 180); /* Teal background */
  }
}
```

### Effects and Animations

Control visual effects:

```css
:root {
  /* Blur effects */
  --blur-sm: blur(4px);
  --blur-md: blur(8px);

  /* Glows */
  --shadow-glow: 0 0 20px oklch(0.65 0.22 285 / 0.3);
}
```

Disable animations by removing the `@layer animations` block or setting:

```css
* {
  animation: none !important;
}
```

## Browser Support

### Modern Browsers (Full Experience)
- Chrome/Edge 111+ (OKLCH, cascade layers, container queries)
- Firefox 113+ (OKLCH, cascade layers)
- Safari 16.4+ (OKLCH, cascade layers)

### Fallback Support
- Older browsers fall back to standard colors and layouts
- All content remains accessible and readable

## CSS Features Reference

### OKLCH Color Space
Perceptually uniform color space that looks consistent across hues. Unlike HSL, OKLCH maintains constant perceived brightness.

**Benefits**:
- Consistent lightness across all hues
- More saturated, vibrant colors
- Better for color manipulation
- Future-proof color technology

### Cascade Layers
Organize CSS by concern rather than specificity wars:

```css
@layer base, components, utilities;

@layer base {
  /* Low priority base styles */
}

@layer components {
  /* Component styles */
}
```

Layers always cascade in declaration order, regardless of selector specificity.

### Container Queries
Elements respond to container size, not viewport:

```css
main {
  container-type: inline-size;
}

@container (min-width: 600px) {
  .entry {
    /* Styles when main container is wide */
  }
}
```

### Fluid Typography
Responsive text without media queries:

```css
/* Scales from 1rem to 1.125rem between viewports */
font-size: clamp(1rem, 0.95rem + 0.25vw, 1.125rem);
```

## Design Philosophy

1. **Lush and Rich**: Dark theme with depth, using layered backgrounds and vibrant accents
2. **Modern Web**: Leverage latest CSS for better DX and performance
3. **Readable First**: High contrast, generous spacing, optimal line-height
4. **Accessible**: Respects user preferences for motion and color schemes
5. **Progressive Enhancement**: Works everywhere, enhanced on modern browsers

## Performance

- **Zero JavaScript**: Pure CSS theme, no runtime overhead
- **Native Features**: Browser-optimized effects (backdrop-filter, gradients)
- **Efficient Selectors**: Modern selectors are faster than complex specificity chains
- **Layer Optimization**: Cascade layers reduce selector complexity

## License

Part of the Rogue Planet project. See the main repository for license details.
