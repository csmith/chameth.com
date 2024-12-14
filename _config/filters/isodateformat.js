import {DateTime} from 'luxon';

export default (date) =>
    DateTime
        .fromJSDate(date)
        .toISODate();