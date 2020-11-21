import * as sentry from "../../sdk/nodejs";

const random = new sentry.Random("my-random", { length: 24 });

export const output = random.result;