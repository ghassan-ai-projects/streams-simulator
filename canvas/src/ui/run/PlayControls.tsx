import { Pause, Play } from "lucide-react";
import { playbackSpeeds } from "../../model/playback";
import styles from "./RunScreen.module.css";

export interface PlayProps {
  readonly playing: boolean;
  readonly speed: number;
  readonly onPlay: () => void;
  readonly onPause: () => void;
  readonly onSpeed: (speed: number) => void;
}

export function PlayControls({ playing, speed, onPlay, onPause, onSpeed }: PlayProps) {
  return (
    <div className={styles.play}>
      <button
        type="button"
        className={styles.playButton}
        aria-keyshortcuts="Space"
        title={playing ? "Pause (Space)" : "Play (Space)"}
        onClick={playing ? onPause : onPlay}
      >
        {playing ? <Pause size={14} aria-hidden="true" /> : <Play size={14} aria-hidden="true" />}
        {playing ? "Pause" : "Play"}
      </button>
      <div className={styles.speeds} role="group" aria-label="Playback speed">
        {playbackSpeeds.map((value) => (
          <button
            key={value}
            type="button"
            className={styles.speed}
            aria-pressed={value === speed}
            onClick={() => {
              onSpeed(value);
            }}
          >
            {`${String(value)}×`}
          </button>
        ))}
      </div>
    </div>
  );
}
