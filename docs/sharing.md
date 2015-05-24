#Sharing data with others
Once you've recorded some time you might want to share your measurements with others. Timeglass uses git-notes and thus stores measurements as metadata in a seperate branch. By default, a "pre-push" hook is installed that automatically pushes your time measurements to the remote whenever you issues a `git push` for regular commits. 

If you disabled this behaviour in the [configuration](/docs/config.md) or if you want to push time data manually you can use:

```
glass push [remote]
```
_NOTE: Remote is optional and defaults to "origin" if not specified._

Similarly, if you wish to retrieve time measurements from the remote you can use the pull command.

```
glass pull [remote]
```

After you've pulled time data from the remote you can happely [query](/docs/query.md) it however you like.

