# Configuration
Timeglass can be configured by creating a `timeglass.json` file in the root of the repository you are tracking. The following example shows all options with their default configuration:

```json
{
	"mbu": "1m",
	"commit_message": " [{{.}}]",
	"auto_push": true
}
```

Alternatively, you can place a `timeglass.json` file with the same structure into any of the following locations to provide default configuration for all projects on the system:

- __OSX:__ /Library/Timeglass
- __Linux:__ /var/lib/timeglass
- __Windows:__ %PROGRAMDATA%\Timeglass or else %ALLUSERSPROFILE%\Timeglass


# Options:
Below are all the possible configuration options, are you interested in configuring anything else? [let us know](https://github.com/Timeglass/glass/issues/7)

## MBU
__key__: `mbu`  

A timer runs in the background and increments by a set amount of time each tick: the "minimal billable unit". It accepts a human readable format that is parsed by: [time.ParseDuration()](http://golang.org/pkg/time/#ParseDuration), e.g: `1h5m2s`

## Commit Message Template
__key__: `commit_message`  

This options allows you to specify how Timeglass should write time measurments to commit messages. To disable this feature completely, provide an empty string, e.g: `"commit_message": ""`

The template is parsed using the standard Go [text/templating](http://golang.org/pkg/text/template/), but you probably only need to know that `{{.}}` is replaced by a human readable representation of the measured time, e.g: `1h5m2s` 

The `.` represents a [time.Duration](http://golang.org/pkg/time/#Duration) struct. This means that you can also call its individual methods from inside the template, for example the following configuration

```"commit_message": " [spent {{printf \"%0.1f\" .Hours}} hours]"```

will output commit messages like this:

```I did something [spent 0.1 hours]```

## Automatically Push Time data
__key__: `auto_push`  
__requirements__: git v1.8.2.1 or higher

Timeglass uses [git-notes](http://git-scm.com/docs/git-notes) for storing commit times. git-notes uses a seperate branch for this data that needs to be explicitely pushed or else data is merely stored local and lost whenever the clone is removed. To prevent this, Timeglass installes a pre-push hook that automatically pushes time data to the same remote as the push itself. If you rather want full control over when to push time data using the `glass push` command, you can disable the automatic behaviour with this options: `"auto_push": false`. The pre-push hook was introduced in git v1.8.2, if you're running an older version the hook is simply not run and this option does nothing.