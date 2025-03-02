# SonicRadio

A stylish TUI radio player making use of [Radio Browser API](https://www.radio-browser.info/) and [Bubbletea](https://github.com/charmbracelet/bubbletea).


## Installation


### Prerequisites

One of the following tools must be installed and available in the PATH:
- Mpv : <https://mpv.io/>
- FFplay : <https://ffmpeg.org/ffplay.html>, comes bundled with ffmpeg
- VLC: <https://www.videolan.org/vlc/>

Use one of the following methods:
- Download one of the available binaries from [Releases](https://github.com/dancnb/sonicradio/releases) page.
- Install using go:

  ```
    go install github.com/dancnb/sonicradio@latest
  ```
- Clone this repository and build from source.

## Usage

After the installation, the command to run the application:

```
    sonicradio
```

Available options:

```
      -debug: creates a log file "sonicradio-[epoch millis].log" in OS specific temp dir
```

![ Demo](demo.gif)

### Keybindings

| Key(s)      |                Action |
| :---------- | --------------------: |
| ↑/k         |                    up |
| │↓/j        |                  down |
| ctrl+f/pgdn |             next page |
| ctrl+b/pgup |             prev page |
| g/home      |           go to start |
| G/end       |             go to end |
| enter       |                  play |
| space       |          pause/resume |
| -           |              volume - |
| +           |              volume + |
| ←/h         |        seek backwards |
| →/l         |          seek forward |
| i           |          station info |
| f           |      favorite station |
| a           |      autoplay station |
| d           |        delete station |
| p/shift+p   | paste deleted station |
| /           |        filter results |
| s           |      open search view |
| #           |  go to station number |
| esc         |     go to now playing |
| shift+tab   |        go to prev tab |
| tab         |        go to next tab |
| v           |           change view |
| ?           |           toggle help |
| q           |                  quit |

## TODO

- [x] Search stations section
- [x] Display rich station information
- [x] Playback history tab
- [x] Settings tab

## License

Sonicradio is licensed under the [MIT License](LICENSE).

### Third-party dependencies

[Bubbletea](https://github.com/charmbracelet/bubbletea/blob/master/LICENSE) MIT License

