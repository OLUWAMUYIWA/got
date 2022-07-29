package pkg

var testConfig string = `
[user]
	name = Pavan Kumar Sunkara ; commenting this out
	email = pavan.sss1991@gmail.com
	username = pksunkara
[init]
	; commenting this out
	defaultBranch = master
[core]
	editor = nvim
	whitespace = fix,-indent-with-non-tab,trailing-space,cr-at-eol
	pager = delta
[sendemail]
	smtpencryption = tls
	smtpuser = pavan.sss1991@gmail.com
	smtppass = password
	smtpserverport = 587
[color "branch"]
	current = yellow bold
	local = green bold
	remote = cyan bold
	; commenting this out
[color "diff"]
	meta = yellow bold
	; commenting this out
	new = green bold
	whitespace = red reverse
[color "status"]
	added = green bold
	line-numbers-right-format = "{np:^4}│ " ; commenting this out
[github]
	user = pksunkara
	token = token
[gitflow "prefix"]
	versiontag = v
[sequence]
	editor = interactive-rebase-tool
[alias]
	a = add --all
	ai = add -i
	ac = apply --check
	#############
	ama = am --abort
	clg = !sh -c 'git clone git://github.com/$1 $(basename $1)' -
	clgp = !sh -c 'git clone git@github.com:$1 $(basename $1)' -
	clgu = !sh -c 'git clone git@github.com:$(git config --get user.username)/$1 $1' -
	#############
	g = grep -p
	opr = !sh -c 'git fo pull/$1/head:pr-$1 && git o pr-$1'
	#############
	pr = prune -v
	#############e1 ; commenting this out
	pl = pull
	pb = pull --rebase
	#############
	plo = pull origin
	ploc = !git pull origin $(git bc)
	pbom = pull --rebase origin master
	pboc = !git pull --rebase origin $(git bc)
	#############
	; commenting this out
	plu = pull upstream
	plum = pull upstream master
	pluc = !git pull upstream $(git bc)
	pbum = pull --rebase upstream master
	pbuc = !git pull --rebase upstream $(git bc)
	#############
	rb = rebase
	remh = reset --mixed HEAD
	resh = reset --soft HEAD
	rehom = reset --hard origin/master
	#############
	r = remote
	ra = remote add
	; commenting this out
	rpu = remote prune upstream
	#############
	rmf = rm -f
	st = !git stash list | wc -l 2>/dev/null | grep -oEi '[0-9][0-9]*'
	#############
	t = tag
	; commenting this out
	wp = show -p--subdirectory-filter $1 master' -
	human = name-rev --name-only --refs=refs/heads/*
[filter "lfs"]
	clean = git-lfs clean -- %f
	smudge = git-lfs smudge -- %f
	process = git-lfs filter-process
	required = true

	x*?
	x+?
	x??
	x{n,m}?
	x{n,}?
	x{n}?

	(re) numbered grouping for submatch
	(?P<name>re) named and numbered submatch
	(?re) non capturin grouping
	

`
