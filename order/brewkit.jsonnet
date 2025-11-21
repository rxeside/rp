local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'order',
];

local proto = [
    'api/server/orderinternalapi/orderinternalapi.proto',
];

project.project(appIDs, proto)