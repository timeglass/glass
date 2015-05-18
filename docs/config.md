# Configuration
Timeglass can be configured by creating a `timeglass.json` file in the root of the repository you are tracking. Timeglass only accepts valid JSON, so if something seems wrong make sure to check your formatting.

## Example
The following example shows all options with their default configuration:

```json
{
	"mbu": "1m",
	"commit_message": " [{{.}}]"
}
```

## MBU
__key__: `mbu`  

A timer runs in the background and increments by set amount of time eacht tick: the "minimal billable unit". It accepts a human readable format that is parsed by: [time.ParseDuration()](http://golang.org/pkg/time/#ParseDuration)

## Commit Message Template
__key__: `commit_message`  

You can specify how you would like write the time you spent to the end of a commit message. To disable this feature completely configure an empty string like this: `"commit_message": ""`

The template is parsed using the standard Go [text/templating](http://golang.org/pkg/text/template/), but you probably only need to know that `{{.}}` is replaced by a human readable format of the measured time.