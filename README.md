[![Go](https://github.com/dhcgn/workplace-sync/actions/workflows/go.yml/badge.svg)](https://github.com/dhcgn/workplace-sync/actions/workflows/go.yml)

# workplace-sync
 
> Keep your tools in sync!

This tool (under heavy development) downloads files from a list of links. These links can be a DNS TXT record or a local *.json file.

So I need only this tool on every of my computers to access easy all my tools.

```
Workplace Sync v0.0.12 (2022-11-22T20:10:06Z go1.19.2)
https://github.com/dhcgn/workplace-sync

host or localSource is required
  -all
        Download all links, except skipped ones
  -checkupdate
        Check for update from github.com
  -host string
        The host which DNS TXT record points to an url of links.json
  -local string
        The local source of links (.json)
  -name string
        The name or preffix of the tool to download
  -update
        Update app with latest release from github.com
  -version
        Return version of app
```

## Optional Powershell Profile

```powershell
# code $PROFILE
if (Test-Path -Path c:\ws\) {
    Get-ChildItem c:\ws\ -Filter *.exe | ForEach-Object{
        $name = $_.Name
        $name = $name.Replace(".exe", "")
        New-Alias -Name $name -Value $_.FullName
    }
}
```

## Demo

![](docs/assets/demo.gif)

## Installation

1. Add DNS TXT record with a link to your JSON file or use my at ws.hdev.io

```json
{
    "links": [
        {
            "url": "https://download.sysinternals.com/files/SysinternalsSuite.zip",
            "version": "latest"
        },
        {
            "name": "age-rc",
            "url": "https://github.com/FiloSottile/age/releases/download/v1.1.0-rc.1/age-v1.1.0-rc.1-windows-amd64.zip",
            "decompress_flat": true,
            "decompress_filter": "\\.exe$",
            "overwrite_files_names": {
                "^age.exe$": "age-rc.exe",
                "^age-keygen.exe$": "age-keygen-rc.exe"
            },
            "version": "v1.1.0-rc.1"
        },
        {
            "name": "BeyondCompare",
            "url": "https://scootersoftware.com/BCompare-4.4.4.27058.exe",
            "type": "installer",
            "version": "4.4.4.27058"
        },
        {
            "name": "ffmpeg",
            "url": "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip",
            "decompress_flat": true,
            "decompress_filter": "\\.exe$",
            "version": "latest",
            "skipped": true
        },        
    ],
    "last_modified": "2022-11-19T22:49:05.718215Z"
}
```

## Usage

The folder `C:\ws\` will be created.

### Manual Selected Download

```
workplace-sync.exe -host ws.hdev.io

Workplace Sync v0.0.10 (2022-11-20T19:47:36Z go1.19.2)
https://github.com/dhcgn/workplace-sync

 INFO  Optain links from DNS TXT record of ws.hdev.io
 SUCCESS  Got 30 links
 INFO  Use download folder c:\ws\
 INFO  The following tools are available:
7z (22.01), BeyondCompare (4.4.4.27058), Everything (1.4.1.1022), ImageMagick (7.1.0-51), OOSU10 (10), Obsidian (1.0.3), SciTE (531), SysinternalsSuite.zip (latest), WinSCP (5.21.5), Wireguard (latest), Wireshark (4.0.1), age (v1.0.0), cloc (v1.94), ffmpeg (latest), git (v2.38.1), jxl (v0.7.0), minisign (0.10), mkcert (latest), mkcert (v1.4.4), nmap (7.93), notepad-plus-plus (v8.4.6), paint.net (4.3.12), putty (latest), puttygen (latest), restic (v0.14.0), upx (4.0.0), vscode (latest), vt-cli (0.10.4), winbox (latest), zstd (v1.5.2)
 INFO  Please select file to download:
>
```
### Pre-Selected Download

```
workplace-sync.exe -host ws.hdev.io -name ag

Workplace Sync v0.0.10 (2022-11-20T19:47:36Z go1.19.2)
https://github.com/dhcgn/workplace-sync

 INFO  Optain links from DNS TXT record of ws.hdev.io
 SUCCESS  Got 30 links
 INFO  Use download folder c:\ws\
 WARNING  No file found, try case-ignore prefix
 SUCCESS  Found file age
age-v1.0.0-windows-amd64.zip 100% |█████████████████████████████████████████████████████| (4.1/4.1 MB, 33 MB/s)
unzip age-keygen.exe 100% |████████████████████████████████████████████████████████████████████| (2/2, 61 it/s)  
```

### Download all files

```
workplace-sync.exe -host ws.hdev.io -all
```

### Future Features

- Integrity check all the files
- Change download location of single files 
- Use a optional config file
- Possiblility of using encrypted files (with https://age-encryption.org/)
- Secure DNS TXT requests
- Update only missing files
- Update only new files
- Allow mutliple DNS TXT records
- Install ps1 script
