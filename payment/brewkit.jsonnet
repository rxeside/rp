local project = import 'brewkit/project.libsonnet';

local appIDs = [
    'payment',
];

local proto = [
    'api/server/paymentinternal/paymentinternal.proto',
];

project.project(appIDs, proto)