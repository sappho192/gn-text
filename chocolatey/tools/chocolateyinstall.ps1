$ErrorActionPreference = 'Stop'

$packageName = 'gn-text'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

$url64 = "https://github.com/sappho192/gn-text/releases/download/v$version$/gn-text_$version$_windows_amd64.zip"

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  checksum64    = '$checksum64$'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
