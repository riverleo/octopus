const express = require('express');
const cors = require('cors');
const cookieParser = require('cookie-parser');
const graphqlHTTP = require('express-graphql');
const build = require('./schema/build');
const shell = require('shelljs');

let schema, ricePort, riceProcess;
const port = 4000, env = process.env.NODE_ENV || 'local';

const getPort = (min=40000, max=49999) => Math.floor(Math.random() * (max - min + 1)) + min;

process.on('SIGINT', () => {
  riceProcess.kill();
  process.exit(0);
});

try {
  (() => {
    ricePort = getPort();
    riceProcess = shell.exec(
      `go run ${__dirname}/main.go -env=${env} -port=${ricePort} -dir=${__dirname}`,
      { async: true },
    );
  })()

  schema = build({ ricePort });
} catch(e) {
  riceProcess.kill();
  throw e;
}

const app = express();
app.use('/', [cors(), cookieParser()], graphqlHTTP({
  schema,
  graphiql: true,
}));

app.listen(port);

console.log(`[${env}] Running a GraphQL server on :${port}.`);
