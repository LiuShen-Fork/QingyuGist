# Custom assets

To add custom assets to your QingyuGist instance, you can use the `$qingyugist-home/custom` directory (where `$qingyugist-home` is the directory where QingyuGist stores its data).

### Logo / Favicon

To add a custom logo or favicon, you can add your own image file to the `$qingyugist-home/custom` directory, then define the relative path in the config.

For example, if you have a logo file `logo.png` in the `$qingyugist-home/custom` directory, you can set the logo path in the config as follows:

#### YAML
```yaml
custom.logo: logo.png
```

#### Environment variable
```sh
export QYG_CUSTOM_LOGO=logo.png
```

Same as the favicon:

#### YAML
```yaml
custom.favicon: favicon.png
```

#### Environment variable
```sh
export QYG_CUSTOM_FAVICON=favicon.png
```

### Instance Name

It is also possible to set a name for your instance, that would be displayed in the title bar instead of 'QingyuGist'.

#### YAML
```yaml
custom.name: My Gists
```

#### Environment variable
```sh
export QYG_CUSTOM_NAME="My Gists"
```
