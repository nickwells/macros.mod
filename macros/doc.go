/*

The macros package can be used to replace parts of a string with the contents
of macro files held in one of a set of directories. You construct the M
object and then call the Substitute method on each string that you wish to
expand. This will find all the sub-strings bracketed with the macro start and
end values and replace them with the macros it knows. If the macro is not
found in the cache and macro directories have been given they are searched
and if a file is found with the same name as the macro (possibly with a
suffix) then the contents of that file are used as the value. Any newly found
macros are cached for further use. Any macro that cannot be found in the
cache or in the macro directories is reported as an error.

Alternatively macro values can be set directly and no macro directories
are needed.

*/
package macros
