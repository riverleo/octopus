const _ = require('lodash');
const { isLeafType, GraphQLObjectType, GraphQLScalarType } = require('graphql');
const { stripType, stripAST, isListType, isPlainListType } = require('./utils');
const { getUserId } = require('../auth');

function parseArgument({ name, value, kind }) {
  var kind = _.get(value, 'kind', kind);
  var parsedValue = _.get(value, 'value', value);

  if (kind === 'IntValue') {
    parsedValue = parseInt(parsedValue);
  } else if (kind === 'FloatValue') {
    parsedValue = parseFloat(parsedValue);
  } else if (kind === 'BooleanValue') {
    parsedValue = (parsedValue === 'true');
  } else if (kind === 'ObjectValue') {
    parsedValue = _.assign(..._.map(value.fields, field => parseArgument(field)));
  } else if (kind === 'ListValue') {
    parsedValue = _.map(value.values, field => parseArgument(field)); }

  if (_.isEmpty(name)) {
    return parsedValue;
  }

  return { [name.value]: parsedValue };
}

function nodeFromAST({ ast, type, args }, variables) {
  if (_.isEmpty(type)) { return; }

  // 쿼리 내부에서 호출되는 구조적 쿼리를 판단합니다.
  var type = type;
  var strippedType = stripType(type);
  const strippedAST = stripAST(ast);

  const isListStruct = _.isPlainObject(type) && type.type instanceof GraphQLObjectType;
  const isPlainListStruct = _.isPlainObject(type) && strippedType instanceof GraphQLObjectType;

  const node = {
    name: strippedAST.name.value,
    type: strippedType.name,
    args: args || argumentsFromAST(strippedAST, variables),
    isLeaf: isLeafType(type),
    isList: isListType(type),
    isPlainList: isPlainListType(type),
  };

  if (isListStruct || isPlainListStruct) {
    type = type.type;
    strippedType = stripType(type);
    node.name = ast.name.value;
    node.type = strippedType.name;
    node.args = argumentsFromAST(ast, variables);
    node.isLeaf = isLeafType(type);
    node.isList = isListType(type);
    node.isPlainList = isPlainListType(type);
  }

  node.fields = _.assign(
    {},
    fieldsFromAST({ ast: stripAST(strippedAST), type: strippedType }, variables),
    fieldsFromReserved({ ast, type }, variables)
  );

  return node;
}

function fieldsFromAST({ ast, type }, variables) {
  const fields = {};

  if (_.isEmpty(ast.selectionSet)) {
    return fields;
  }

  _.forEach(ast.selectionSet.selections, (ast) => {
    const node = nodeFromAST({ ast, type: type._fields[ast.name.value] }, variables);

    if (_.isEmpty(node)) { return; }

    fields[ast.name.value] = node;
  });

  return fields;
}

function fieldsFromReserved({ ast, type }, variables) {
  const fields = {};
  const strippedType = stripType(type, true);
  var selections = _.get(ast, 'fieldNodes[0].selectionSet.selections');

  if (_.isEmpty(selections)) {
    selections = _.get(ast, 'selectionSet.selections');
  }

  _.forEach(selections, (ast) => {
    const name = ast.name.value;

    if (!_.startsWith(name, '_') || _.isEqual(name, '_data')) {
      return;
    }

    const node = nodeFromAST({ ast, type: strippedType._fields[name] }, variables);

    if (_.isEmpty(node)) { return; }

    fields[name] = node;
  });

  return fields;
}

function argumentsFromAST(ast, variables) {
  return _.assign(..._.map(ast.arguments, ({ name, value }) => {
    var kind = _.get(value, 'kind', kind);
    var parsedValue = _.get(value, 'value', value);

    if (kind === 'IntValue') {
      parsedValue = parseInt(parsedValue);
    } else if (kind === 'FloatValue') {
      parsedValue = parseFloat(parsedValue);
    } else if (kind === 'BooleanValue') {
      parsedValue = (parsedValue === 'true');
    } else if (kind === 'ObjectValue') {
      parsedValue = _.assign(..._.map(value.fields, field => parseArgument(field)));
    } else if (kind === 'ListValue') {
      parsedValue = _.map(value.values, field => parseArgument(field));
    } else if (kind === 'Variable') {
      parsedValue = variables[parsedValue.name.value];
    }

    if (_.isEmpty(name)) {
      return parsedValue;
    }

    return { [name.value]: parsedValue };
  }));
}

function createRequest(req, { ast, type, args }) {
  const name = _.get(ast.operation.name, 'value', 'anonymous');
  const operation = ast.operation.operation;
  const node = nodeFromAST({ ast, type, args }, ast.variableValues);

  return { name, operation, node, userId: getUserId(req) };
}

module.exports = createRequest;
