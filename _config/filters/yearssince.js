import {DateTime} from 'luxon';

export default (date) =>
    Math.floor(DateTime.fromJSDate(date).diffNow('years').years * -1);