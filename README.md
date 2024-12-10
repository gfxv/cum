# CUM (Comprehensive Universal Mapper)

## Building
To build **CUM** from source:
1) If not installed, install **go 1.23 or newer** on your machine. You can get **go** from [the official website](https://go.dev/doc/install).
2) If not installed, install **Taskfile**. You can get **Taskfile** from [the official website](https://taskfile.dev/installation/).
3) Clone this repo: `git clone https://github.com/gfxv/cum.git`
4) In the `cum` directory run `task` or `task build`, which will result in a binary file named `cum`.

## Usage
### Configuration
To have access to **CUM** from any place in your terminal, you need to add **CUM** to PATH.

To do it, follow these steps:
1. Add the following line to your shell configuration file (`~/.bashrc`, `~/.zshrc`, etc.), replacing `<path-to-binary>` with the actual path to compiled binary:
```sh
export PATH="<path-to-binary>:$PATH"
```
2. Reload shell configuration or restard the terminal.

Here is example of how you can reload shell configuration for `~/.zshrc`:
```sh
source ~/.zshrc
```
### Using CUM
Example of encoding input file named `input.txt` into mp4 named `out.mp4`:
```
cum -a encode -in input.txt -out out.mp4
```

Example of decode input file named `out.mp4` into original file name `decoded.txt`:
```
cum -a decode -in out.mp4 -out decoded.txt
```


All available commands for **CUM**:
```
$ cum -help
Usage of ./cum:
  -a string
    	Specifies action to perform: encode or decode. Encode will take an input and convert it to mp4 format. Decode will attempt to convert provided video to oiginal format
  -in string
    	Path to input file
  -out string
    	Path to output file (will be created if not exists or overwritten if already exists)
  -qrisize int
    	Define a size of QR Code. Can be omitted (default 1024)
```