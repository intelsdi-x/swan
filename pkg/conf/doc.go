/*
Package conf extends builtin 'flag' packagea to provide:
- environment parsing with predefined prefix,
- config file generation with grouping (instead of lexicographical order),
- ability to extract current values of or registered flags (defined with wrappers),
- new types of flags e.g. SliceFlag,
- predefined flags for logging (logrus integration),
*/
package conf
