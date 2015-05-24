#Querying Time data
_Timeglass_ attaches the time you spent as metadata to Git commits. This means you have the full power of Git at your disposal when it comes to querying for time data. It also means that it always consists of two steps: 

1. First, select the work you're interested in by fetching a list of commit hashes (seperated by newlines) from Git using either `git rev-list` or `git log --prety=%H`.
2. Second, pipe this list into `gass sum` to add all time entries together. It will output thet total time in a human readable format (e.g 1h59m10s)

Because querying Git can be a science in it own right we included some common patterns below. Have question about your data that isn't answers by any of the examples below? [let us know](https://github.com/timeglass/glass/issues/9)

## How much time was spent on...

##### ...on all commits since "yesterday"?
	git log --since="1 days ago" --pretty=%H | glass sum

##### ...on all commits authored by "advanderveer" since "this morning"?
	git log --author=advanderveer --since="9am" --pretty=%H | glass sum

##### ...on all commits since "May 20"?
	git log --since="may 20" --pretty=%H | glass sum

##### ...all commits up up to and including the current HEAD?
	git rev-list --all | glass sum

##### ...all commits authored by "advanderveer" since tag "v0.5.0"?
	git rev-list --author=advanderveer v0.5.0..HEAD | glass sum

##### ...all commits authored by "advanderveer" up to an including current HEAD?
	git rev-list --all --author=advanderveer | glass sum

##### ...commits in the current branch (given the current branch is not master)?
	git rev-list master..HEAD | glass sum

##### ...on commits of the branch that were merged in commit "d2192a058"
	git rev-list d2192a058^..d2192a058 | glass sum

NOTE: To show all merges that occured: `git log --merges --format=oneline`



