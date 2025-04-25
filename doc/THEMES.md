# Customizing Themes

The News Aggregator supports both light and dark themes. You can customize these themes by modifying the CSS files or adding your own styles.

## Location of Theme Files

The theme-related styles are located in the following file:

```
web/templates/static/main.css
```

## Customizing Light Theme

To customize the light theme, look for the `:root` class in the CSS file and modify the properties as needed. For example:

```css
:root {
  --body-bg: #fff;
  --panel-bg: #f3f3f3;
  --border: 1px solid #c8c8c8;
  --text-color: #616161;
  --text-color-active: #ffffff;
  --hover-bg: #BD2A2E;
  --hover-link: #980E2F;
  --count-bg: #fff;
  --box-shadows: 1px 1px 7px 0px var(--hover-link);
}
```

## Customizing Dark Theme

To customize the dark theme, look for the `data-theme="dark"` class in the CSS file and modify the properties as needed. For example:

```css
[data-theme="dark"] {
  --body-bg: #121212;
  --panel-bg: #1e1e1e;
  --border: 1px solid #333333;
  --text-color: #e0e0e0;
  --text-color-active: #ffffff;
  --hover-bg: #bb86fc;
  --hover-link: #bb86fc;
  --count-bg: #1e1e1e;
  --box-shadows: 1px 1px 7px 0px var(--hover-link);
}
```