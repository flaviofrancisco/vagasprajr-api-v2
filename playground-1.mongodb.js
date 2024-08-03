use('vagasprajrdb');

db.users.updateMany({}, { $set: { is_email_confirmed: true } });