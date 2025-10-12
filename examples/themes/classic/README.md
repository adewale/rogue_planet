# Classic Planet Theme

A faithful recreation of the classic Planet theme inspired by [Planet Venus](https://github.com/adewale/venus) `classic_fancy` theme, as seen on sites like [Planet Mozilla](https://planet.mozilla.org/) and [OpenStreetMap Blogs](https://blogs.openstreetmap.org/).

## Features

- **Classic Planet Look**: Lowercase headers with negative letter spacing
- **Right Sidebar**: Subscriptions list with RSS feed icons
- **Blue Color Scheme**: Purple headers (#200080) and blue boxes (#a0c0ff)
- **Bitstream Vera Sans Font**: The iconic Planet font family
- **Minimal Design**: Text-focused, content-first approach
- **RSS Feed Icons**: Orange RSS icons next to each subscription
- **Responsive**: Adapts to mobile screens

## Preview

This theme recreates the aesthetic of classic Planet aggregators from the 2000s-2010s era:

- Main content on the left with a black border on the right
- Fixed-width sidebar (220px) on the right side
- Entry titles in blue boxes
- Date groupings in purple
- Small RSS icons inline with feed names

## Installation

### Option 1: Copy to Your Planet Directory

```bash
# Copy the entire classic theme to your planet
cp -r examples/themes/classic ~/myplanet/theme-classic/

# Update config.ini to use the theme
vim ~/myplanet/config.ini
```

Add to your `config.ini`:

```ini
[planet]
template = ./theme-classic/template.html
```

### Option 2: Use Absolute Path

```ini
[planet]
template = /full/path/to/rogue_planet/examples/themes/classic/template.html
```

## Usage

1. **Configure your planet** with the classic theme:

```bash
cd ~/myplanet
vim config.ini
```

```ini
[planet]
name = my planet
link = https://planet.example.com
owner_name = Your Name
template = ./theme-classic/template.html
group_by_date = true    # Recommended for classic look
```

2. **Generate your planet**:

```bash
rp update
```

Static assets (CSS and SVG files) are automatically copied from the theme's `static/` folder to `public/static/` when you generate your site.

3. **View the result**:

```bash
open public/index.html
```

## Customization

### Colors

Edit `static/style.css` in your theme folder to customize the color scheme:

```css
/* Header color */
h1 {
    color: #808080;  /* Change site title color */
}

/* Date group headers */
h2 {
    color: #200080;  /* Change to your preferred color */
}

/* Entry title boxes */
h3 {
    background-color: #a0c0ff;  /* Light blue background */
    border: 1px solid #5080b0;  /* Border color */
}
```

### Fonts

The theme uses Bitstream Vera Sans (the classic Planet font). To change:

```css
body, h1, h2, h3, h4 {
    font-family: "Your Font", "Fallback Font", sans-serif;
}
```

### Sidebar Width

To adjust the sidebar width:

```css
body {
    margin-right: 220px;  /* Match sidebar width */
}

.sidebar {
    width: 220px;  /* Adjust width */
}
```

## Differences from Default Theme

| Feature | Default Theme | Classic Theme |
|---------|---------------|---------------|
| Layout | Modern flexbox | Fixed sidebar on right |
| Typography | System fonts | Bitstream Vera Sans |
| Colors | Blue links | Purple headers, blue boxes |
| Headers | Standard case | Lowercase with spacing |
| Feed Icons | None | Orange RSS icons |
| Style | Modern, clean | Nostalgic, classic |

## Compatibility

- ✅ Works with all Rogue Planet features
- ✅ Supports date grouping
- ✅ Shows feed health in sidebar
- ✅ Responsive on mobile
- ✅ Content Security Policy compliant
- ✅ No external dependencies

## Theme Structure

The classic theme includes:
- `template.html` - Main HTML template
- `static/style.css` - External CSS stylesheet (automatically copied to `public/static/`)
- `static/feed-icon.svg` - RSS feed icon (automatically copied to `public/static/`)
- `LICENSE` - Attribution and licensing information
- `README.md` - This documentation

## Notes

- CSS and images are external files in the `static/` folder
- Static assets are automatically copied to `public/static/` when generating
- The sidebar uses absolute positioning (classic Planet style)
- On mobile, the sidebar moves below the content

## Credits and Attribution

**Theme Design:**
Based on the [Venus](https://github.com/rubys/venus) `classic_fancy` theme by Sam Ruby and contributors. Colors and typography inspired by Planet Mozilla, Planet Debian, and other classic Planet sites.

**RSS Feed Icon:**
The RSS feed icon is from the [Mozilla Foundation](https://commons.wikimedia.org/wiki/File:Feed-icon.svg), tri-licensed under:
- Mozilla Public License Version 1.1
- GNU General Public License (GPL)
- GNU Lesser General Public License (LGPL)

See [LICENSE](LICENSE) file for complete attribution and licensing details.

## See Also

- [Rogue Planet README](../../../README.md) - Main documentation
- [Default Theme](../../../pkg/generator/generator.go) - Modern default theme
- [Venus Themes](https://github.com/adewale/venus/tree/master/themes) - Original theme inspiration
