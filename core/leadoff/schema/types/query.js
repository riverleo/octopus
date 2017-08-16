const _ = require('lodash');
const {
  GraphQLInt,
  GraphQLFloat,
  GraphQLString,
  GraphQLBoolean,
  GraphQLID,
  GraphQLList,
  GraphQLSchema,
  GraphQLNonNull,
  GraphQLEnumType,
  GraphQLObjectType,
  GraphQLInputObjectType,
} = require('graphql');
const { camelize } = require('underscore.string');
const { stripType } = require('../utils');
const { GraphQLDateTime, GraphQLErrorType } = require('./customs');
const { isCompositeType } = require('graphql/type');

// ========================================================
// Query
// ========================================================

function createQueryType(type, many=false) {
  const lastIndexOfTypeName = _.size(type.name) - 1;
  const name = `${_.lowerFirst(type.name)}${!many ? '' : 'List'}`;
  const parsedType = !many ? type : new GraphQLNonNull(
    new GraphQLObjectType({
      name: _.upperFirst(name),
      fields: () => ({
        _total: { type: new GraphQLNonNull(GraphQLInt) },
        _count: { type: new GraphQLNonNull(GraphQLInt) },
        _limit: { type: new GraphQLNonNull(GraphQLInt) },
        _offset: { type: new GraphQLNonNull(GraphQLInt) },
        _data: { type: new GraphQLNonNull(new GraphQLList(type)) },
        _error: { type: GraphQLErrorType },
      }),
    })
  );

  return {
    [name]: {
      type: parsedType,
    },
  };
}

function createQuery(types) {
  const fields = {};

  _.forEach(types, type => {
    const list = createQueryType(type, true);
    const single = createQueryType(type);

    _.assign(fields, list, single);
  });

  return new GraphQLObjectType({
    name: 'Query',
    fields,
  });
};

// ========================================================
// Arguments
// ========================================================

function createWhereInputType(schema, originType, targetType) {
  const type = targetType || originType;

  const fields = {
    eq: { type, name: 'eq' },
    ne: { type, name: 'ne' },
    in: { type: new GraphQLList(type), name: 'in' },
    notIn: { type: new GraphQLList(type), name: 'notIn' },
    nil: { type: GraphQLBoolean, name: 'nil' },
  };

  switch(originType.name) {
    case 'Int':
      fields.gt = { type: GraphQLInt, name: 'gt' };
      fields.gte = { type: GraphQLInt, name: 'gte' };
      fields.lt = { type: GraphQLInt, name: 'lt' };
      fields.lte = { type: GraphQLInt, name: 'lte' };
      break;
    case 'Float':
      fields.gt = { type: GraphQLFloat, name: 'gt' };
      fields.gte = { type: GraphQLFloat, name: 'gte' };
      fields.lt = { type: GraphQLFloat, name: 'lt' };
      fields.lte = { type: GraphQLFloat, name: 'lte' };
      break;
    case 'String':
      fields.like = { type: GraphQLString, name: 'like' };
      fields.ilike = { type: GraphQLString, name: 'ilike' };
      break;
  }

  const inputName = `${type.name}WhereInput`;
  const inputType = new GraphQLInputObjectType({
    name: inputName,
    fields: () => fields,
  });

  inputType._fields = fields;
  schema._typeMap[inputName] = inputType;

  return inputType;
}

function createScalarWhereInputTypes(schema) {
  return _.reduce(
    [GraphQLInt, GraphQLFloat, GraphQLString, GraphQLBoolean, GraphQLID, GraphQLDateTime],
    (col, scalarType) => {
      const inputType = createWhereInputType(schema, scalarType);

      col[scalarType.name] = inputType;
      schema._typeMap[inputType.name] = inputType;

      return col;
  }, {});
}

function createWhereInputTypes(schema, types) {
  const scalarWhereInputTypes = createScalarWhereInputTypes(schema);
  const whereInputTypes = _.reduce(types, (col, type) => {
    const whereInputFields = _.reduce(type._fields, (col, fieldType) => {
      if (_.startsWith(fieldType.name, '_')) { return col; }

      const mergedFieldType = schema._typeMap[type.name]._fields[fieldType.name];
      const actualType = scalarWhereInputTypes[fieldType.type.name];

      schema._typeMap[type.name]
      col[fieldType.name] = {
        type: actualType,
        name: fieldType.name,
      };

      return col;
    }, { _object: { name: '_object', type: GraphQLBoolean, defaultValue: true } });

    const inputName = `${type.name}WhereInput`;
    const inputType = new GraphQLInputObjectType({
      name: inputName,
      fields: () => whereInputFields,
    });

    inputType._fields = whereInputFields;
    col[type.name] = inputType;
    schema._typeMap[inputName] = inputType;

    return col;
  }, {});

  _.forEach(whereInputTypes, (whereInputType, typeName) => {
    const mergedType = schema._typeMap[typeName];

    if (!mergedType || typeName === '_object') { return; }

    _.forEach(mergedType._fields, (fieldType) => {
      if (!isCompositeType(fieldType.type)) { return; }
      if (_.startsWith(fieldType.name, '_')) { return; }

      const strippedType = stripType(fieldType.type);

      whereInputType._fields[camelize(strippedType.name, true)] = {
        type: whereInputTypes[strippedType.name],
        name: camelize(strippedType.name, true),
      };
    });
  });

  return whereInputTypes;
}

function createOrderInputTypes(schema, types) {
  const orderDirection = new GraphQLEnumType({
    name: 'OrderDirection',
    values: {
      ASC: { value: 'ASC' },
      DESC: { value: 'DESC' },
    },
  });

  const orderFunction = new GraphQLEnumType({
    name: 'OrderFunction',
    values: {
      SUM: { value: 'SUM' },
    },
  });

  const orderInputFields = {
    to: {
      name: 'to',
      type: new GraphQLNonNull(orderDirection),
    },
    func: {
      name: 'func',
      type: orderFunction,
    },
  };

  const orderInputType = new GraphQLInputObjectType({
    name: 'OrderInput',
    fields: () => orderInputFields,
  });

  orderInputType._fields = orderInputFields;
  schema._typeMap['OrderInput'] = orderInputType;
  schema._typeMap['OrderFunction'] = orderFunction;
  schema._typeMap['OrderDirection'] = orderDirection;

  const orderInputTypes = _.reduce(types, (col, type) => {
    const orderInputFields = _.reduce(type._fields, (col, fieldType) => {
      if (_.startsWith(fieldType.name, '_')) { return col; }

      col[fieldType.name] = {
        type: orderInputType,
        name: fieldType.name,
      };

      return col;
    }, { _object: { name: '_object', type: GraphQLBoolean, defaultValue: true } });

    const inputName = `${type.name}OrderInput`;
    const inputType = new GraphQLInputObjectType({
      name: inputName,
      fields: () => orderInputFields,
    });

    inputType._fields = orderInputFields;
    col[type.name] = inputType;
    schema._typeMap[inputName] = inputType;

    return col;
  }, {});

  _.forEach(orderInputTypes, (orderInputType, typeName) => {
    const mergedType = schema._typeMap[typeName];

    if (!mergedType || typeName === '_object') { return; }

    _.forEach(mergedType._fields, (fieldType) => {
      if (!isCompositeType(fieldType.type)) { return; }
      if (_.startsWith(fieldType.name, '_')) { return; }

      const strippedType = stripType(fieldType.type);

      orderInputType._fields[fieldType.name] = {
        type: orderInputTypes[strippedType.name],
        name: camelize(strippedType.name, true),
      };
    });
  });

  return orderInputTypes;
}

function updateQueryFromSchema(schema, types) {
  schema._typeMap.DateTime = GraphQLDateTime;
  const whereInputTypes = createWhereInputTypes(schema, types);
  const orderInputTypes = createOrderInputTypes(schema, types);

  _.forEach(schema._queryType._fields, (query, queryName) => {
    const strippedType = stripType(query.type);
    const whereInputType = whereInputTypes[strippedType.name];
    const orderInputType = orderInputTypes[strippedType.name];

    _.forEach(strippedType._fields, (field) => {
      if (field.type.name === 'DateTime') {
        field.type = GraphQLDateTime;
      }
    });

    query.args = query.args.concat([
      {
        name: '_where',
        type: whereInputType,
      },
      {
        name: '_or',
        type: new GraphQLList(whereInputType),
      },
      {
        name: '_and',
        type: new GraphQLList(new GraphQLList(whereInputType)),
      },
      {
        name: '_order',
        type: orderInputType,
      },
      {
        name: '_offset',
        type: GraphQLInt,
      },
      {
        name: '_limit',
        type: GraphQLInt,
      },
    ]);
  });
}

exports.createQuery = createQuery;
exports.updateQueryFromSchema = updateQueryFromSchema;
