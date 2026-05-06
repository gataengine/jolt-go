#ifndef SETTINGS_H
#define SETTINGS_H
#include <stdint.h>
#include <stdbool.h>

typedef struct JPH_Vec3 { float x, y, z; } JPH_Vec3;
typedef struct JPH_Quat { float x, y, z, w; } JPH_Quat;
typedef struct JPH_RVec3 { float x, y, z; } JPH_RVec3;
typedef struct JPH_Shape JPH_Shape;
typedef uint16_t JPH_ObjectLayer;
typedef uint8_t JPH_MotionType;

typedef struct JPH_BodyCreationSettings {
    JPH_RVec3       Position;
    JPH_Quat        Rotation;
    JPH_Vec3        LinearVelocity;
    JPH_Vec3        AngularVelocity;
    uint64_t        UserData;
    JPH_ObjectLayer ObjectLayer;
    JPH_MotionType  MotionType;
    bool            AllowSleeping;
    float           Friction;
    float           Restitution;
    float           LinearDamping;
    float           AngularDamping;
    JPH_Shape*      Shape;
} JPH_BodyCreationSettings;

void JPH_BodyCreationSettings_Init(JPH_BodyCreationSettings*);

#endif
