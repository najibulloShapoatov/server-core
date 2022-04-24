// Package version implements semantic version according to semver.org 2.0.0 specs.
// Beside regular version parsing, the package can also process and compare version ranges
//
//   v, err := version.New("1.2.3-alpha")
//   v1, err := version.New("^1.2.0")
//
//   // version formats accepted
//   v.Equal(v1) // true
//   v.Equal("v2.0.0") // false
//   v.Equal("2.0.0") // false
//   v.Equal("1.2.3-alpha+sh1.123") // false
//   v.Equal("1.2.3+sh1.123") // false
//   v.Equal("1.2.3-beta") // false
//   // Range inclusion comparison
//   v.Equal("~1.2.5") // true
//   v.Equal("^1.0") // true
//   v.Equal("*") // true
//
//   // version comparison methods
//   v.Equal("1.2.3-alpha") // true
//   v.LessThan("1.2.3-beta") // true
//   v.LessEqThan("1.3.0") // true
//   v.GreaterThan("0.0.1") // true
//   v.GreaterEqThan("2.0.0") // true
//
//   // Release methods
//   v, _ := version.New("1.2.3-alpha+tag")
//   v.ReleaseDev("beta") // 1.2.3-beta
//   v.ReleasePatch() // 1.2.4
//   v.ReleaseMinor() // 1.3.0
//   v.ReleaseMahor() // 2.0.0
package version
