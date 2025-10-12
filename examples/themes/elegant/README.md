# Elegant Theme

A modern, refined theme for Rogue Planet featuring a clean design system with CSS custom properties, responsive layout, and contemporary typography.

## Features

- **Modern Design System**: Built with CSS custom properties for easy customization
- **Responsive Layout**: Mobile-first design that adapts to all screen sizes
- **Typography Scale**: Modular scale for consistent text hierarchy
- **Semantic Colors**: Thoughtful color palette with accent, success, warning, and error states
- **Layout Primitives**: Flexible container, stack, and cluster layouts
- **Print Optimized**: Clean print styles for paper output

## Theme Structure

```
elegant/
├── template.html        # HTML template with semantic markup
├── static/
│   └── style.css       # Complete design system and styles
└── README.md           # This file
```

Static assets (CSS files) are automatically copied from the theme's `static/` folder to `public/static/` when you generate your site.

## Usage

### Basic Setup

1. Copy the elegant theme to your project:
```bash
mkdir -p themes
cp -r examples/themes/elegant themes/
```

2. Update your `config.ini`:
```ini
[planet]
name = My Planet
link = https://planet.example.com
owner_name = Your Name
template = ./themes/elegant/template.html
```

3. Generate your site:
```bash
rp generate
```

## Customization

### Design Tokens

The theme uses CSS custom properties defined in `static/style.css`. Edit these to customize the look:

```css
:root {
  /* Typography scale */
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  --text-base: 1rem;
  --text-lg: 1.25rem;
  --text-xl: 1.5rem;

  /* Spacing rhythm */
  --space-3: 1rem;
  --space-4: 1.5rem;
  --space-5: 2rem;

  /* Color palette */
  --gray-50: #fafafa;
  --gray-900: #18181b;
  --accent: #2563eb;

  /* Layout */
  --max-width: 72rem;
  --radius: 0.375rem;
}
```

### Colors

- **Neutral Palette**: `--gray-50` through `--gray-900` for text and backgrounds
- **Accent**: `--accent` for links and interactive elements (default: blue)
- **Semantic**: `--success`, `--warning`, `--error` for status indicators

### Typography

- **Sans Serif**: System font stack optimized for each platform
- **Monospace**: For code blocks and technical content
- **Modular Scale**: Six text sizes from `--text-xs` to `--text-2xl`

### Layout

- **Max Width**: Content constrained to `--max-width` (72rem / 1152px)
- **Sidebar**: Fixed width (320px) on desktop, full width on mobile
- **Spacing**: Consistent rhythm using `--space-*` variables

### Responsive Behavior

- **Desktop (≥768px)**: Two-column layout with sidebar
- **Mobile (<640px)**: Single column, reduced spacing, smaller text
- **Print**: Sidebar hidden, optimized for paper

## Design Philosophy

The Elegant theme follows modern web design principles:

1. **Progressive Enhancement**: Works on all devices, enhanced on capable ones
2. **Semantic HTML**: Proper use of article, header, aside, footer elements
3. **Accessibility**: High contrast ratios, logical heading hierarchy
4. **Performance**: Minimal CSS, system fonts, no external dependencies
5. **Maintainability**: CSS custom properties make theming straightforward

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers (iOS Safari, Chrome Mobile)

## License

Part of the Rogue Planet project. See the main repository for license details.
