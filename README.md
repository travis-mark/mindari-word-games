# Mindari's Word Games

Mindari's Word Games is a tool to extract Wordle, etc... scores from a shared Discord channel. I named the tool after a bard I played at a D&D oneshot. Discord's docs suggested having a cute mascot, so I kept it around. 

Usage:

        mindari <command> [arguments]

The commands are:

        bot         Run discord bot for slash commands
        list        List channels with data
        help        Show this list
        monitor     Periodically monitor for posted scores
        rescan      Do a full rescan of a channel (in case of defects or edits)
        serve       Start a local webserver to show stats and a leaderboard
        stats       Print stats to standard output to use for custom graphs
        update      Scan all channels from their most recent entry forward

For those hooking into the live version, just invite the bot to your channel from the site. Commands are not needed, but will come shortly.

For those looking to self-host a private version, clone this repo and run `go build` followed by `./mindari serve`.
