#ifndef TYPES_H
#define TYPES_H

#include <stdint.h>

typedef struct JPH_Shape JPH_Shape;          // opaque
typedef struct JPH_BodyID { uint32_t id; } JPH_BodyID;

typedef enum JPH_MotionType {
    JPH_MotionType_Static = 0,
    JPH_MotionType_Kinematic = 1,
    JPH_MotionType_Dynamic = 2,
} JPH_MotionType;

typedef uint32_t JPH_ObjectLayer;

#endif
