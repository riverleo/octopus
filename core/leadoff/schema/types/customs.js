const {
  GraphQLInt,
  GraphQLList,
  GraphQLString,
  GraphQLNonNull,
  GraphQLObjectType,
  GraphQLScalarType,
} = require('graphql');

exports.GraphQLDateTime = new GraphQLScalarType({
  name: 'DateTime',
  description: 'The `DateTime` scalar type represents ISO8601',
  serialize: (value) => {
    return value;
  },
  parseValue: (value) => {
    return value;
  },
  parseLiteral: (ast) => {
    if (ast.kind === Kind.String) {
      return value;
    }
    return null;
  }
});

exports.GraphQLErrorType = new GraphQLObjectType({
  name: 'Error',
  fields: () => ({
    _count: { type: new GraphQLNonNull(GraphQLInt) },
    _data: {
      type: new GraphQLList(
        new GraphQLObjectType({
          name: 'ErrorData',
          fields: () => ({
            key: { type: new GraphQLNonNull(GraphQLString) },
            code: { type: new GraphQLNonNull(GraphQLInt) },
            message: { type: new GraphQLNonNull(GraphQLString) },
          }),
        })
      ),
    },
  }),
});
