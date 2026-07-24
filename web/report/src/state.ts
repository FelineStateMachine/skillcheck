export const initialState = { selectedDifference: '', selectedSide: 'left' };
export function reduce(state, action) {
  if (action.type === 'difference') return { ...state, selectedDifference: action.id };
  if (action.type === 'side') return { ...state, selectedSide: action.side };
  return state;
}
