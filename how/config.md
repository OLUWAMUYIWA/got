## Config

To learn about the basics of regular expressions, you may want to look at my distillation of [re2](https://github.com/google/re2/wiki/Syntax), found [here](https://github.com/OLUWAMUYIWA/field_gleanings/blob/main/regex.md). The original source would be more helpful. Russ Cox (I think he authored it) did a great job.


In creating a parser for `git`'s config, I used `regular expressions`.
Now, first I would like to explain the structure of `git`'s `config` file, after which 