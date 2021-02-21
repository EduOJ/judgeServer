# EduOJJudger

A golang judge server for EduOJ.

Using [Judger](https://github.com/suntt2019/Judger).

## Requirements

Here are the requirements for the scripts used to judge.  
Follow the requirements, and refer to the [example](/example).

> Note: The way of writing `compile` files in the samples is not necessary, it's just to test that the compilation process does exist.  
>Just write `#!/bin/bash` in `compile` file, and write the run script in the `run` file in use.

### Requirements for all the scripts
* stored as a zip file with a file named `compile`
* `compile` has executing permission
* `compile` creates a file named `run` with execute permission in the same directory
* could not have a file named `updated_at`

### Requirements for build script of any language
* Arguments:
    * First arg: path to the run directory
    * Following args: build args of the problem
* Path:
    * The script isn't executed at the run directory, remember to output files in the run directory
        * eg: using build argument `-o $1/a.out`
    * The code is placed in the run directory, named `code`
    * The build result should be placed in the run directory
* Permission:
    * The build script is executed by `root`
    * The run directory is set to mode `0777`
    * The runnable file generated should be executable by any user
* Input & Output:
    * No input for build scripts
    * All the output will be stored and be displayed to the submitter, please don't output any private content  
    * Error output will be written to the same file as normal output
* Exit code:
    * Success: exit with `0`
    * Failed: exit with any non-zero number
    
### Requirements for run script of any language
* Arguments:
    * Only one arg: path to the run directory
* Path:
    * The script isn't executed at the run directory, please use the run directory to indicate paths of build result files
        * eg: using `$1/a.out` to indicate executable file in run directory
* Permission:
    * The run script is executed by `root`
    * The run directory is set to mode `0777`
    * The file that will be executed should be executable by any user
* Input & Output:
    * No input for run scripts
    * Output a run command:
        * The executing path is the string before the first space
        * The arguments are split by spaces after the first space
    * The progress of running this command by run user will be regard as running progress of this run
    * Error output will be discard
* Exit code:
    * Success: exit with `0`
    * Failed: exit with any non-zero number

### Requirements for compare script of any problem
* Arguments:
    * First arg: path to the output file of the user program
    * Second arg: path to the standard output file
    * Third arg: run directory
    * Fourth arg: path to the input file
* Permission:
    * The compare script is executed by `root`
* Path:
    * The script isn't executed at the run directory, please use the run directory to indicate paths of other files
        * eg: using `$3/other_file` to indicate a file in the run directory
* Input & Output:
    * No input for compare scripts
    * The output will be stored and only displayed for users with permission like administers
    * Error output will be written to the same file as normal output
* Exit code:
    * Accepted: exit with `0`
    * Wrong answer: exit with `1`
    * Presentation error: exit with `2`
    * Failed: exit with other number
