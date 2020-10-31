package main

const EXIT_SUCCESS = 0

const EXIT_CONFIG_READ_ERROR = 11
const EXIT_FILESYSTEM_ACCESS_ERROR = 12

const EXIT_ERRONEOUS_ADD_ARGS = 21
const EXIT_ERRONEOUS_CRYPT_ARGS = 22

const EXIT_GIT_ERROR = 31

const EXIT_CONFIG_WRITE = 41

const EXIT_ERROR_INTERNAL = 99

//----------------------------------------------------

const CONFIG_PATH = "~/.config/gogitmirror.toml"

const PROGNAME = "goGitmirror"
const PROGVERSION = "0.4"

const TEMPFOLDERNAME = "gogitmirror"
const NETRCPATH = "~/.netrc"

const SALT = "iBl0Vf3SPGq65m4X"

//----------------------------------------------------

const STAT_COL_NAME = 28
const STAT_COL_BRANCH = 28
const STAT_COL_SOURCE = 8
const STAT_COL_LOCAL = 8
const STAT_COL_TARGET = 8
