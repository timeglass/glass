#Querying Time data


### How much time was spent on all commits up up to an including current HEAD?
`git rev-list --all | glass sum`

### How much time was spent on all commits authored by "advanderveer" up to an including current HEAD?
`git rev-list --all --author=advanderveer | glass sum`

### How much time was spent on the current branch (given the current branch is not master)?
`git rev-list master..HEAD | glass sum`

### How much time was spent on the branch that was merged in commit "d2192a058"
`git rev-list d2192a058^..d2192a058 | glass sum`

NOTE: To show all merges that occured: `git log --merges --format=oneline`



### How much total time spent on all commits since tag "0.5"?
`<coming soon>`

### How much time was spent on all commits authored by  "advanderveer" since tag "0.5"?
`<coming soon>`

### How much time was spent on all commits authored by "advanderveer" since "15 may 2015"?
`<coming soon>`