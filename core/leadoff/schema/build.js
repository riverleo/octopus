const _ = require('lodash');
const fs = require('fs');
const axios = require('axios');
const { parse } = require('graphql');
const {
  printSchema,
  GraphQLSchema,
} = require('graphql');
const { mergeSchema, updateQuery } = require('./merge');
const createMock = require('./mock');
const createRequest = require('./request');
const { octopusPort } = require('../graphql');
const { types, query, scalars } = require('./types');
const { updateQueryFromSchema } = require('./types/query');

module.exports = ({ octopusPort }) => {
  var schema = new GraphQLSchema({ types, query });

  const customSchemaPath = `${__dirname}/../custom.graphql`;
  if (fs.existsSync(customSchemaPath)) {
    const customSchema = fs.readFileSync(customSchemaPath, 'utf8');
    schema = mergeSchema(schema, customSchema);
  }

  updateQueryFromSchema(schema, types);

  fs.writeFileSync(`${__dirname}/../.schema`, printSchema(schema), 'utf8');

  _.forEach(schema._typeMap.Query._fields, (field) => {
    field.resolve = (parent, args, req, ast) => {
      const type = field.type;
      const request = createRequest(req, { ast, type, args });
      const output = axios.post(`http://localhost:${octopusPort}`, request, { headers: req.headers }).then(res => res.data);

      if (output) {
        return output;
      }

      return createMock(node, args);
    };
  });

  _.forEach(scalars, (scalar) => _.assign(schema._typeMap[scalar.name], scalar));

  return schema;
}
