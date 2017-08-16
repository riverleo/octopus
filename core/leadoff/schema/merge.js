const fs = require('fs');
const _ = require('lodash');
const { camelize } = require('underscore.string');
const {
  parse,
  buildASTSchema,
  printSchema,
  GraphQLArgument,
  GraphQLInputObjectType,
} = require('graphql');
const { GraphQLList, GraphQLString } = require('graphql');
const { isInputType, isCompositeType } = require('graphql/type');
const { stripType, isListType, isPlainListType } = require('./utils');
const { GraphQLDateTime } = require('./types/customs');

// documentA가 documentB로 덮어씌워집니다.
function mergeDocument(A, B) {
  _.forEach(B.definitions, (b) => {
    const f = (a) => a.name.value === b.name.value && a.kind === b.kind;
    const a = _.find(A.definitions, f);

    // 신규로 생성된 정의인 경우 추가합니다.
    if (_.isNil(a)) {
      A.definitions.push(b);
      return;
    }

    mergeArrayByName('fields', a, b);
    mergeArrayByName('interfaces', a, b);
    mergeArrayByName('directives', a, b);
  });

  return A;
}

function mergeArrayByName(name, A, B) {
  _.forEach(B[name], (b) => {
    const a = _.find(A[name], (a) => a.name.value === b.name.value);

    if (_.isNil(a)) {
      A[name].push(b);
      return;
    }

    _.assign(a, b);
  });
}

// schemaA가 schemaB로 덮어씌워집니다.
function mergeSchema(schemaA, schemaB) {
  var documentA = typeof(schemaA) === 'string' ? parse(schemaA) : parse(printSchema(schemaA));
  const documentB = typeof(schemaB) === 'string' ? parse(schemaB) : parse(printSchema(schemaB));

  return buildASTSchema(mergeDocument(documentA, documentB));
}

exports.mergeSchema = (...schemas) => {
  var mergedSchema;

  _.forEach(schemas, (schema) => {
    if (_.isEmpty(schema)) { return; }

    if (_.isEmpty(mergedSchema)) {
      mergedSchema = schema;
      return;
    }

    try {
      mergedSchema = mergeSchema(mergedSchema, schema);
    } catch(e) {
      console.warn('fail to build custom.graphql\n');
    }
  });

  return mergedSchema;
};
