import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api';

export const schoolHandlers = [
  http.get(`${BASE}/schools`, () => {
    return HttpResponse.json([
      { id: 'sch-1', name: 'Duke' },
      { id: 'sch-2', name: 'UNC' },
    ]);
  }),
];
