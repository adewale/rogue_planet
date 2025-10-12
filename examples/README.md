# Rogue Planet Examples

This directory contains example configuration files and themes for Rogue Planet.

## Files

- **config.ini** - Complete configuration example with all options documented
- **feeds.txt** - Sample feed list with popular blogs and sites
- **themes/** - Pre-built themes for your planet

## Usage

```bash
# Create a folder for your own planet
mkdir ~/myplanet
cd ~/myplanet

# Initialize
rp init

# Copy example config
cp examples/config.ini config.ini
vim config.ini

# Add feeds
rp add-feed https://blog.golang.org/feed.atom
rp add-feed https://github.blog/feed/

# Update and generate
rp update
```

## Themes

> **📚 Complete theme guide: [../THEMES.md](../THEMES.md)**

Rogue Planet includes five themes:

### Default Theme
The built-in responsive theme. No configuration needed - works out of the box.
- Modern flexbox layout
- System font stack
- Feed sidebar with health status
- Mobile-friendly responsive design

### Classic Theme
A faithful recreation of the classic Planet Venus theme:
- Right sidebar with RSS feed icons
- Bitstream Vera Sans typography
- Blue/purple color scheme (#200080, #a0c0ff)
- Lowercase headers with negative letter spacing
- Nostalgic 2000s-2010s Planet aesthetic

See [themes/classic/README.md](themes/classic/README.md) for details.

### Elegant Theme
A modern, refined theme with sophisticated typography:
- CSS custom properties design system
- Georgia serif body text for readability
- Modular typography scale
- Generous spacing and elegant color palette
- Responsive layout with sticky sidebar
- Print-optimized styles

See [themes/elegant/README.md](themes/elegant/README.md) for details.

### Dark Theme
A lush, modern dark theme showcasing cutting-edge CSS:
- OKLCH color space for vibrant, perceptually uniform colors
- CSS cascade layers for maintainable architecture
- Fluid typography with clamp() functions
- Glassmorphism effects with backdrop-filter
- Container queries and modern selectors (`:has()`, `:where()`)
- Electric accent colors (cyan, violet, pink, emerald)

See [themes/dark/README.md](themes/dark/README.md) for details.

### Flexoki Theme
An inky color scheme with automatic light/dark mode switching:
- Designed for reading with warm paper tones
- Automatic light/dark mode (respects system preference)
- Oklab color space for perceptual consistency
- Serif typography optimized for prose
- 8 vibrant accent colors
- No JavaScript required

See [themes/flexoki/README.md](themes/flexoki/README.md) for details.

## Creating Your Own Theme

See the [Complete Theme Guide](../THEMES.md) for:
- Template variables reference (40+ variables)
- Template functions (formatDate, relativeTime, etc.)
- Theme creation tutorial
- Customization examples
- Security best practices

## Customization

All example files are heavily commented. Customize `config.ini` with your planet's name, domain, and preferences.

## See Also

- [QUICKSTART.md](../QUICKSTART.md) - 5-minute setup guide
- [WORKFLOWS.md](../WORKFLOWS.md) - Complete operational workflows
- [README.md](../README.md) - Full documentation
