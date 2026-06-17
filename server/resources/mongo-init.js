const adminDb = db.getSiblingDB("admin")

console.log("[mongo-init] create user")

adminDb.createUser({
    user: process.env.MONGO_INITDB_ROOT_USERNAME,
    pwd: process.env.MONGO_INITDB_ROOT_PASSWORD,
    roles: [
        {
            role: 'root',
            db: "admin",
        },
        {
            role: 'dbOwner',
            db: process.env.MONGO_INITDB_DATABASE,
        },
    ],
});
console.log("[mongo-init] user created")

adminDb.createUser({
    user: process.env.MONGO_METRICS_USERNAME,
    pwd: process.env.MONGO_METRICS_USER_PASSWORD,
    roles: [
        { role: "clusterMonitor", db: "admin" },
        { role: "readAnyDatabase", db: "admin" },
        { role: "read", db: "local" }
    ]
});
console.log("[mongo-init] metrics user created")
