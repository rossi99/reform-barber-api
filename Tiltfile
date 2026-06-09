# -- Extensions ---
load('ext://dotenv', 'dotenv')

# -- Env ---
dotenv()

# -- Docker ---
docker_compose('./docker-compose.yml')
dc_resource('postgres', labels=['infra'])
dc_resource('pgadmin', labels=['infra'], links=[link('http://localhost:5050', 'pgAdmin')])
dc_resource('minio', labels=['infra'], links=[link('http://localhost:9001', 'MinIO console')])
dc_resource('minio-init', labels=['infra'], resource_deps=['minio'])
dc_resource('mailpit', labels=['infra'], links=[link('http://localhost:8025', 'Mailpit')])

# -- Sever ---
local_resource('api',
    serve_cmd='go run ./cmd/server',
    deps=['cmd', 'internal', 'db', 'go.mod', 'go.sum'],
    ignore=['**/*_test.go'],
    resource_deps=['postgres'],
    links=[link('http://localhost:8080/api/barbers', 'API smoke')],
    labels=['backend'],
)

# -- Database ---
db_label=["database"]
local_resource('migration up',
    labels=db_label,
    cmd='task migrate-up',
    auto_init=False,
)

local_resource('migration down',
    labels=db_label,
    cmd='task migrate-down',
    auto_init=False,
)

local_resource('migration check',
    labels=db_label,
    cmd='task migrate-check',
    auto_init=False,
)

# -- UI ---
local_resource('web',
    serve_cmd='npm run dev',
    serve_dir='../web',
    resource_deps=['api'],
    links=[link('http://localhost:3000', 'Web')],
    labels=['frontend'],
)
