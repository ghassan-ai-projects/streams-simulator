import type { ReviewSession, ReviewState } from "../../../app/review";
import { shared } from "../../primitives";

export function LiveControls({ review, state }: { readonly review: ReviewSession; readonly state: ReviewState }) {
  return (
    <div role="group" aria-label="Live following">
      <span role="status">
        {state.following ? "Live · following newest record" : "Following paused · viewing history"}
      </span>{" "}
      {!state.following && (
        <button
          type="button"
          className={shared.primary}
          onClick={() => {
            review.backToLive();
          }}
        >
          Back to live
        </button>
      )}
    </div>
  );
}
