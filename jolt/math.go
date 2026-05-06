package jolt

import "math"

// Vec3 is a 3-component single-precision vector. Layout matches JPH_Vec3.
type Vec3 struct{ X, Y, Z float32 }

func (v Vec3) Add(o Vec3) Vec3      { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }
func (v Vec3) Sub(o Vec3) Vec3      { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }
func (v Vec3) Scale(s float32) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }
func (v Vec3) Neg() Vec3            { return Vec3{-v.X, -v.Y, -v.Z} }

func (v Vec3) Dot(o Vec3) float32 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }

func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{
		X: v.Y*o.Z - v.Z*o.Y,
		Y: v.Z*o.X - v.X*o.Z,
		Z: v.X*o.Y - v.Y*o.X,
	}
}

func (v Vec3) LengthSq() float32 { return v.Dot(v) }
func (v Vec3) Length() float32   { return float32(math.Sqrt(float64(v.LengthSq()))) }

func (v Vec3) Normalized() Vec3 {
	l := v.Length()
	if l == 0 {
		return Vec3{}
	}
	return v.Scale(1 / l)
}

// Quat is a single-precision unit quaternion. Layout matches JPH_Quat.
type Quat struct{ X, Y, Z, W float32 }

func QuatIdentity() Quat { return Quat{W: 1} }

func (q Quat) Mul(o Quat) Quat {
	return Quat{
		X: q.W*o.X + q.X*o.W + q.Y*o.Z - q.Z*o.Y,
		Y: q.W*o.Y - q.X*o.Z + q.Y*o.W + q.Z*o.X,
		Z: q.W*o.Z + q.X*o.Y - q.Y*o.X + q.Z*o.W,
		W: q.W*o.W - q.X*o.X - q.Y*o.Y - q.Z*o.Z,
	}
}

func (q Quat) Conjugate() Quat { return Quat{-q.X, -q.Y, -q.Z, q.W} }

func (q Quat) Inverse() Quat {
	n := q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
	if n == 0 {
		return Quat{}
	}
	inv := 1 / n
	c := q.Conjugate()
	return Quat{c.X * inv, c.Y * inv, c.Z * inv, c.W * inv}
}

func (q Quat) RotateVec3(v Vec3) Vec3 {
	// v' = q * (0,v) * q^-1, simplified for unit quats
	u := Vec3{q.X, q.Y, q.Z}
	s := q.W
	return u.Scale(2 * u.Dot(v)).Add(
		v.Scale(s*s - u.Dot(u)).Add(
			u.Cross(v).Scale(2 * s),
		),
	)
}
