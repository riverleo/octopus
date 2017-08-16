const _ = require('lodash');
const fs = require('fs');
const path = require('path');
const colors = require('colors/safe');

const filenames = [];
fs.readdirSync(__dirname).forEach((filename) => {
  if (!_.endsWith(filename, '.json')) { return; }

  const name = filename.split('.json')[0];
  filenames.push(name);
});

function search(candidate, config={ lang: 'ko' }) {
  const filename = _.find(filenames, (target) => {
    const transformedCandidate = _.snakeCase(candidate);

    if (transformedCandidate === target) {
      return target;
    } else if (_.endsWith(transformedCandidate, target)) {
      return target;
    } else if (_.startsWith(transformedCandidate, target)) {
      return target;
    }
  });

  if (_.isNil(filename)) { return; }

  try {
    const data = require(`./${filename}.json`);
    return _.sample(data[config.lang]);
  } catch(e) {
    console.warn(colors.red(`${colors.bold(`${filename}.json`)} is not formatted correctly or no data for ${colors.bold(config.lang)}. Please confirm.`));
    return;
  }
}

function getRandomInt(field, min=1, max=10000) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function getRandomFloat(field, min=1, max=1000000) {
  return Math.random() * (max - min) + min;
}

function getRandomString(field) {
  const candidate = field.name;
  const string = search(candidate);

  if (string) {
    return string;
  } else {
    return `No suitable ${candidate} data was found.`;
  }
}

function getRandomBoolean(field) {
  return true;
}

function getRandomDateTime(field) {
  const seed = getRandomInt(null, -2000000000, 2000000000);
  return new Date(new Date().getTime() + seed).toISOString();
}

function getMock(node, args) {
  const mock = {};

  _.forEach(node.fields, (field) => {
    switch (field.type) {
      case 'Int':
        mock[field.name] = field.name === 'id' && args.id ? args.id : getRandomInt(field);
        break;
      case 'Float':
        mock[field.name] = getRandomFloat(field);
        break;
      case 'String':
        mock[field.name] = getRandomString(field);
        break;
      case 'Boolean':
        mock[field.name] = getRandomBoolean(field);
        break;
      case 'DateTime':
        mock[field.name] = getRandomDateTime(field);
        break;
      default:
        mock[field.name] = getMock(field);
    }
  });

  return mock;
};

function createMock(node, args) {
  if (!node.isList) {
    return getMock(node, args);
  }

  const limit = _.get(node, 'args.limit', 10);
  const offset = _.get(node, 'args.offset', 0);

  return _.map(_.range(limit), (index) => getMock(node, index));
};

module.exports = createMock;
