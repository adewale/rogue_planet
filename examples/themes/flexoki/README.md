# Flexoki Theme

An inky color scheme for Rogue Planet inspired by analog inks and warm paper tones. Designed by [Steph Ango](https://stephango.com/flexoki), adapted for feed aggregation with automatic light/dark mode switching.

## Features

### Color Philosophy
- **Warm Paper Tones**: Background inspired by cream and warm paper shades
- **Inky Accents**: Eight vibrant accent colors derived from Oklab color space
- **Perceptual Balance**: Colors maintain consistent brightness relationships
- **Analog Inspiration**: Emulates the behavior of physical ink and paper

### Design Features
- **Automatic Light/Dark Mode**: Respects your system's color scheme preference
- **No JavaScript Required**: Pure CSS implementation using `prefers-color-scheme`
- **Optimized for Reading**: Serif typography with generous line-height (1.7-1.8)
- **High Contrast**: WCAG AA compliant in both modes
- **Warm Aesthetic**: Reduces eye strain with paper-like background tones

### Technical Features
- **Oklab Color Space**: Scientifically accurate perceptual color matching
- **CSS Custom Properties**: Easy customization of all colors
- **Responsive Layout**: Mobile-friendly design
- **Print Optimized**: Clean print styles included
- **Accessibility**: Focus indicators, reduced motion support

## Preview

### Light Mode
- Background: Warm paper (#FFFCF0)
- Text: Deep black (#100F0F)
- Links: Deep blue (#205EA6)
- Accents: 600-level colors (darker, more saturated)

### Dark Mode
- Background: Deep black (#100F0F)
- Text: Warm gray (#CECDC3)
- Links: Bright blue (#4385BE)
- Accents: 400-level colors (lighter, softer)

### Accent Colors

The theme uses 8 accent colors for various UI elements:

| Color | Light Mode | Dark Mode | Usage |
|-------|------------|-----------|-------|
| Red | #AF3029 | #D14D41 | Errors, feed failures |
| Orange | #BC5215 | #DA702C | Code, special highlights |
| Yellow | #AD8301 | #D0A215 | Warnings |
| Green | #66800B | #879A39 | Success indicators |
| Cyan | #24837B | #3AA99F | Link hover states |
| Blue | #205EA6 | #4385BE | Primary links |
| Purple | #5E409D | #8B7EC8 | Date headers |
| Magenta | #A02F6F | #CE5D97 | Special elements |

## Installation

### Quick Setup

1. Copy the theme to your planet directory:
```bash
mkdir -p themes
cp -r examples/themes/flexoki themes/
```

2. Update your `config.ini`:
```ini
[planet]
name = My Planet
link = https://planet.example.com
owner_name = Your Name
template = ./themes/flexoki/template.html
group_by_date = true    # Recommended for date-grouped layout
```

3. Generate your site:
```bash
rp generate
```

4. View the result:
```bash
open public/index.html
```

The theme will automatically switch between light and dark modes based on your system preference!

## Customization

### Testing Light/Dark Modes

**macOS:**
- System Preferences → General → Appearance → Light/Dark
- Command + Shift + L in Safari (developer mode)

**Windows:**
- Settings → Personalization → Colors → Choose your color

**Browser DevTools:**
- Chrome/Edge: DevTools → Command Menu (Cmd/Ctrl+Shift+P) → "Render" → Emulate CSS media feature prefers-color-scheme
- Firefox: DevTools → Settings → Enable dark theme simulation

### Changing Colors

Edit `static/style.css` to customize the color palette:

```css
:root {
  /* Light mode colors */
  --link: var(--blue);          /* Primary link color */
  --link-hover: var(--cyan);    /* Link hover color */
  --success: var(--green);      /* Success indicators */
  --error: var(--red);          /* Error indicators */
}

@media (prefers-color-scheme: dark) {
  :root {
    /* Dark mode inherits same semantic colors */
    /* But accent colors are automatically lighter (400 level) */
  }
}
```

### Typography

The theme uses a serif font stack for readability:

```css
--font-serif: 'Iowan Old Style', 'Palatino Linotype', 'URW Palladio L', P052, serif;
```

To change to a different font:

```css
:root {
  --font-serif: Georgia, 'Times New Roman', serif;
  /* or */
  --font-serif: Charter, 'Bitstream Charter', 'Sitka Text', Cambria, serif;
}
```

### Layout Adjustments

Adjust sidebar width:

```css
.sidebar {
  width: 260px;  /* Change width */
}

.layout {
  gap: 60px;     /* Adjust spacing between main and sidebar */
}
```

### Custom Accent Color Mapping

Change which accent colors are used for specific elements:

```css
h2 {
  color: var(--purple);  /* Change to any accent: red, orange, yellow, etc. */
}

.feed-error {
  color: var(--red);     /* Error color */
}
```

## Color System Reference

### Base Colors (Same in Light/Dark)

```css
--paper: #FFFCF0     /* Lightest value */
--black: #100F0F     /* Darkest value */
```

### Text Colors

**Light Mode:**
```css
--tx: #100F0F        /* Primary text */
--tx-2: #1C1B1A      /* Secondary text */
--tx-3: #282726      /* Tertiary text */
```

**Dark Mode:**
```css
--tx: #CECDC3        /* Primary text */
--tx-2: #B7B5AC      /* Secondary text */
--tx-3: #9C9A92      /* Tertiary text */
```

### Background Colors

**Light Mode:**
```css
--bg: #FFFCF0        /* Primary (paper) */
--bg-2: #F2F0E5      /* Secondary */
```

**Dark Mode:**
```css
--bg: #100F0F        /* Primary (black) */
--bg-2: #1C1B1A      /* Secondary */
```

### UI Colors

**Light Mode:**
```css
--ui: #E6E4D9        /* UI elements */
--ui-2: #DAD8CE      /* UI hover */
--ui-3: #CECDC3      /* UI active */
```

**Dark Mode:**
```css
--ui: #282726        /* UI elements */
--ui-2: #343331      /* UI hover */
--ui-3: #403E3C      /* UI active */
```

## Browser Support

### Full Support (Light + Dark)
- Chrome/Edge 76+ (2019)
- Firefox 67+ (2019)
- Safari 12.1+ (2019)
- All modern mobile browsers

### Fallback Support
- Older browsers display light mode only
- All content remains fully accessible

## Design Philosophy

Flexoki is designed around three core principles:

1. **Legibility First**: High contrast, generous spacing, optimal line-height
2. **Warm & Natural**: Paper-like tones reduce eye strain
3. **Perceptual Consistency**: Oklab color space maintains brightness relationships

The color scheme works equally well for:
- Long-form reading
- Code syntax highlighting
- UI elements and data visualization

## Comparison with Other Themes

| Theme | Light/Dark | Color Philosophy | Typography |
|-------|------------|------------------|------------|
| **Flexoki** | ✅ Automatic | Warm, inky, analog | Serif |
| Classic | ❌ Light only | Cool blues, nostalgic | Bitstream Vera |
| Elegant | ❌ Light only | Neutral grays | Georgia serif |
| Dark | ❌ Dark only | Vibrant OKLCH | Sans-serif |

## Accessibility

- ✅ **WCAG AA Contrast**: All text meets 4.5:1 ratio in both modes
- ✅ **Focus Indicators**: Visible keyboard navigation
- ✅ **Reduced Motion**: Respects `prefers-reduced-motion`
- ✅ **Semantic HTML**: Proper heading hierarchy
- ✅ **Print Friendly**: Clean print stylesheet

## Theme Structure

```
flexoki/
├── template.html         # HTML template with color-scheme meta tag
├── static/
│   ├── style.css        # CSS with automatic light/dark switching
│   └── feed-icon.svg    # RSS feed icon
├── README.md            # This file
└── LICENSE              # MIT License with attribution
```

## Attribution

**Flexoki Color Scheme** designed by [Steph Ango](https://stephango.com/flexoki)
- Website: https://stephango.com/flexoki
- GitHub: https://github.com/kepano/flexoki
- License: MIT

**Rogue Planet Theme Adaptation** by Rogue Planet contributors
- Based on Flexoki color values
- Implements automatic light/dark mode switching
- Optimized for feed aggregation and reading

## Credits

- **Color Design**: Steph Ango
- **Original Flexoki**: https://github.com/kepano/flexoki
- **Theme Adaptation**: Rogue Planet project
- **RSS Icon**: Mozilla Foundation (MPL/GPL/LGPL tri-license)

## Further Reading

- [Flexoki Homepage](https://stephango.com/flexoki)
- [Oklab Color Space](https://bottosson.github.io/posts/oklab/)
- [Rogue Planet Themes Guide](../../../THEMES.md)

## License

This theme adaptation is part of the Rogue Planet project.

The Flexoki color scheme is MIT licensed by Steph Ango.

See [LICENSE](LICENSE) for complete license information and attribution.
