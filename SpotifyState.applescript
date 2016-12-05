#!/usr/bin/env osascript

if application "Spotify" is frontmost then
	return "foreground"
end if
if application "Spotify" is running then
	tell application "Spotify"
		return player state as string  -- stopped/playing/paused
	end tell
else
	return "closed"
end if
