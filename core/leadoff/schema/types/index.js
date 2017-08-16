const _ = require('lodash');
const {
  GraphQLID,
  GraphQLInt,
  GraphQLList,
  GraphQLFloat,
  GraphQLString,
  GraphQLBoolean,
  GraphQLNonNull,
  GraphQLObjectType,
} = require('graphql');
const db = require('../../db');
const { createQuery } = require('./query');
const { camelize, classify } = require('underscore.string');
const { GraphQLDateTime, GraphQLErrorType } = require('./customs');

function objectTypeFromTable({ name, columns }) {
  const fields = _.assign({}, ..._.map(columns, (column) => scalarTypeFromColumn(column)));
  fields._error = { type: GraphQLErrorType };

  return new GraphQLObjectType({
    name: classify(name),
    fields: () => fields,
  });
}

function scalarTypeFromColumn({ name, type, key }, isNonNil=false) {
  var parsedType = GraphQLString;

  if (key === 'PRI') {
    parsedType = GraphQLID;
  } else if (_.startsWith(type, 'int')) {
    parsedType = GraphQLInt;
  } else if (_.startsWith(type, 'float')) {
    parsedType = GraphQLFloat;
  } else if (type === 'tinyint(1)') {
    parsedType = GraphQLBoolean;
  } else if (type === 'datetime' || type === 'timestamp') {
    parsedType = GraphQLDateTime;
  }

  if (type === 'datetime' || type === 'timestamp') {
    return {
      [camelize(name)]: {
        type: parsedType,
        args: {
          format: { type: GraphQLString },
        },
      },
    };
  }

  return { [camelize(name)]: { type: parsedType } };
}

function createTypes(tables) {
  const types = [];

  _.forEach(tables, table => {
    types.push(objectTypeFromTable(table));
  });

  return types;
}

exports.types = createTypes(db.tables);
exports.query = createQuery(exports.types);
