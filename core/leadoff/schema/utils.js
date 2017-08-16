const _ = require('lodash');
const { GraphQLList } = require('graphql');

function isListType(type) {
  if (_.has(type, '_fields._data')) {
    return true;
  }

  if (!_.has(type, 'ofType')) {
    return false;
  }

  return isListType(type.ofType);
}

function isPlainListType(type) {
  if (type instanceof GraphQLList) {
    return true;
  }

  if (!_.has(type, 'ofType')) {
    return false;
  }

  return isPlainListType(type.ofType);
}

function stripAST(ast) {
  if (_.has(ast, 'fieldNodes')) {
    return ast.fieldNodes[0];
  }

  var returnAST = ast;
  _.forEach(_.get(ast, 'selectionSet.selections'), (selection) => {
    if (selection.name.value === '_data') {
      returnAST = selection;
    }
  });

  return returnAST;
}

function stripType(type, isPlainStrip=false) {
  const ctype = _.get(type, '_fields._data.type');

  if (!isPlainStrip && ctype && _.endsWith(type.name, 'List')) {
    return ctype.ofType.ofType
  }

  if (_.has(type, 'type')) {
    return stripType(type.type, isPlainStrip);
  }

  if (!_.has(type, 'ofType')) {
    return type;
  }

  return stripType(type.ofType, isPlainStrip);
}

exports.stripAST = stripAST;
exports.stripType = stripType;
exports.isListType = isListType;
exports.isPlainListType = isPlainListType;
