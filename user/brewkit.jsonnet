local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'user',
];

local proto = [
    'api/server/userpublicapi/userpublicapi.proto',
];

project.project(appIDs, proto)