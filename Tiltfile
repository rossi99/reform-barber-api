docker_compose('./docker-compose.yml')

dc_resource('postgres', labels=['infra'])
dc_resource('pgadmin', labels=['infra'], links=[link('http://localhost:5050', 'pgAdmin')])
dc_resource('minio', labels=['infra'], links=[link('http://localhost:9001', 'MinIO console')])
dc_resource('minio-init', labels=['infra'], resource_deps=['minio'])
dc_resource('mailpit', labels=['infra'], links=[link('http://localhost:8025', 'Mailpit')])

api_env = {
    'DATABASE_URL': 'postgres://postgres:postgres@localhost:5432/reform_barber?sslmode=disable',
    'JWT_SECRET': 'dev-secret-change-me',
    'PORT': '8080',
    'STORAGE_BACKEND': 'local',
    'UPLOADS_DIR': './uploads',
}

local_resource(
    'api',
    serve_cmd='go run ./cmd/server',
    serve_env=api_env,
    deps=['cmd', 'internal', 'db', 'go.mod', 'go.sum'],
    ignore=['**/*_test.go'],
    resource_deps=['postgres'],
    links=[link('http://localhost:8080/api/barbers', 'API smoke')],
    labels=['backend'],
)

local_resource(
    'web',
    serve_cmd='npm run dev',
    serve_dir='../web',
    resource_deps=['api'],
    links=[link('http://localhost:3000', 'Web')],
    labels=['frontend'],
)
