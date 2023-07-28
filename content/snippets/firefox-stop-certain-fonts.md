---
title: Stop using certain local fonts
group: Firefox
---

Some lighter fonts are barely readable on lower contrast backgrounds, and some web designers like to err on the side of stupidly light fonts. To prevent Firefox from using these, create a new fontconfig file in e.g. `~/.config/fontconfig/firefox.conf` containing:

```xml
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
	<include>fonts.conf</include>
	<selectfont>
		<rejectfont>
			<glob>*/DejaVuSans-ExtraLight.ttf</glob>
		</rejectfont>
	</selectfont>
	<selectfont>
		<rejectfont>
			<glob>*/Inconsolata-ExtraLight.ttf</glob>
		</rejectfont>
	</selectfont>
	<selectfont>
		<rejectfont>
			<glob>*/Inconsolata-Light.ttf</glob>
		</rejectfont>
	</selectfont>	
</fontconfig>
```

Then start Firefox with the `FONTCONFIG_FILE=~/.config/fontconfig/firefox.conf` environment variable set (e.g. by [adding it to the desktop entry](https://wiki.archlinux.org/title/Desktop_entries#Modify_environment_variables))