const _ = require('lodash');
const jwt = require('jsonwebtoken');
const secret = 'AgHy(Y%f6y9cPxCo/R#mc6TKMqgR4oAk';  // Be careful not to spill outside.

exports.getUserId = (req) => {
  // returns the user ID.
  const header = req.headers['authorization'];
  const cookie = req.cookies['authorization'];

  let type, token;
  if (header || cookie) {
    const auth = (header || cookie).split(' ');
    type = _.lowerCase(auth[0]);
    token = auth[1];
  }

  if (type === 'jwt') {
    try {
      const decoded = jwt.verify(token, secret);
      return decoded.user_id;
    } catch(e) {}
  }

  return;
}
