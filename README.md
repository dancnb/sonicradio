# SonicRadio

A stylish TUI radio player making use of [Radio Browser API](https://www.radio-browser.info/) and [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Prerequisites
* Go : https://go.dev/doc/install
* MPV  : https://mpv.io/

## Installation

    go install github.com/dancnb/sonicradio@latest

## Usage

After the installation, the command to run the application:

    sonicradio


![ Demo](demo.gif)

### Keybindings

| Key(s)          |                          Action |
|:----------------|--------------------------------:|
|↑/k              |                              up |
│↓/j              |                            down |
|ctrl+f/pgdn      |                       next page |
|ctrl+b/pgup      |                       prev page |
|g/home           |                     go to start |
|G/end            |                       go to end |
|enter            |                            play |
|space            |                    pause/resume |
|-/<              |                        volume - |
|+/>              |                        volume + |
|←/h              |                  seek backwards |
|→/l              |                    seek forward |
|i                |                    station info |
|f                |                favorite station |
|d                |                  delete station |
|p/shift+p        |           paste deleted station |
|/                |                  filter results |
|s                |                open search view |
|#                |            go to station number |
|esc              |               go to now playing |
|shift+tab        |                  go to prev tab |
|tab              |                  go to next tab |
|?                |                     toggle help |
|q                |                            quit |

## TODO

- [x] Search stations section
- [x] Display rich station information
- [ ] Playback history tab
- [ ] Settings tab

## License

Sonicradio is licensed under the [MIT License](LICENSE).

### Third-party dependencies

[Bubbletea](https://github.com/charmbracelet/bubbletea/blob/master/LICENSE) MIT License