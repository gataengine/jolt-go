package jolt

// Vec3 is a 3-component single-precision vector. Layout matches JPH_Vec3.
type Vec3 struct{ X, Y, Z float32 }

func (v Vec3) Add(o Vec3) Vec3      { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }
func (v Vec3) Sub(o Vec3) Vec3      { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }
func (v Vec3) Scale(s float32) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }
func (v Vec3) Neg() Vec3            { return Vec3{-v.X, -v.Y, -v.Z} }
