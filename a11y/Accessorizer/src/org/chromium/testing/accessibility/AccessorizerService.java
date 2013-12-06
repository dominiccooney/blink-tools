package org.chromium.testing.accessibility;

import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.Writer;
import java.lang.Thread;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.ArrayList;
import java.util.List;

import android.accessibilityservice.AccessibilityService;
import android.graphics.Rect;
import android.os.AsyncTask;
import android.util.Log;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityNodeInfo;
//import android.view.accessibility.AccessibilityRecord;

import org.json.JSONException;
import org.json.JSONStringer;

public class AccessorizerService extends AccessibilityService {
	private static final String LOG_TAG = "AccessorizerService";

	public AccessorizerService() {
		messages = new ArrayList<String>();
		lock = messages;
	}
	
    @Override
    public void onServiceConnected() {
    	Log.d(LOG_TAG, "onServiceConnected");
    	
    	Thread t = new Thread() {
    		@Override
    		public void run() {
    			try {
    			    serverSocket = new ServerSocket(9000);
    			    
    			    while (true) {
    			    	final Socket newSocket = serverSocket.accept();
    			    	
    			    	synchronized (lock) {
    			    		Log.d(LOG_TAG, "setting up socket");
    			    		
    			    		if (writer != null) {
				    			try {
				    				writer.close();
				    			} catch (IOException e) {
				    				Log.e(LOG_TAG, e.toString());
				    			}
				    			writer = null;
    			    		}
			    			
			    			try {
			    				writer = new OutputStreamWriter(newSocket.getOutputStream());
			    				newSocket.setKeepAlive(true);
			    				Log.d(LOG_TAG, "set up new socket");
			    			} catch (IOException e) {
			    				Log.e(LOG_TAG, e.toString());
			    			}
    			    	}
    			    }
    			} catch (IOException e) {
    				Log.e(LOG_TAG, e.toString());
    			}
    		}
    	};
    	t.setDaemon(true);
    	t.start();
    	
    	t = new Thread() {
    		@Override
    		public void run() {
    			synchronized (lock) {
    				while (true) {
    					// Wait for messages
    					while (messages.size() == 0) {
    						try {
    							lock.wait();
    						} catch (InterruptedException e) {
    							Log.e(LOG_TAG, e.toString());
    						}
    					}
    					
    					// If there's no connected sink, drop them on the floor
    					if (writer == null) {
    						Log.d(LOG_TAG, "dropping " + messages.size() + " messages, no connected socket");
    						messages.clear();
    						continue;
    					}

    					// Send messages
    					try {
	    					for (String message : messages) {
	    						writer.write(message);
	    					}
	    					writer.flush();
    					} catch (IOException e) {
    						Log.e(LOG_TAG, e.toString());
    					}
    					messages.clear();
    				}
    			}
    		}
    	};
    	t.setDaemon(true);
    	t.start();
    }
    
    @Override
    public void onInterrupt() {
    	Log.d(LOG_TAG, "onInterrupt");
    }

    @Override
    public void onDestroy() {
    	super.onDestroy();
    	Log.d(LOG_TAG, "onDestroy");
    	close();
    }
    
    private static String eventTypeToString(int eventType) {
    	switch (eventType) {
    	case AccessibilityEvent.TYPE_ANNOUNCEMENT:
    		return "Announcement";
    	case AccessibilityEvent.TYPE_GESTURE_DETECTION_END:
    		return "Gesture Detection End";
    	case AccessibilityEvent.TYPE_GESTURE_DETECTION_START:
    		return "Gesture Detection Start";
    	case AccessibilityEvent.TYPE_NOTIFICATION_STATE_CHANGED:
    		return "Notification State Changed";
    	case AccessibilityEvent.TYPE_TOUCH_EXPLORATION_GESTURE_END:
    		return "Touch Exploration Gesture End";
    	case AccessibilityEvent.TYPE_TOUCH_EXPLORATION_GESTURE_START:
    		return "Touch Exploration Gesture Start";
    	case AccessibilityEvent.TYPE_TOUCH_INTERACTION_END:
    		return "Touch Interaction End";
    	case AccessibilityEvent.TYPE_TOUCH_INTERACTION_START:
    		return "Touch Interaction Start";
    	case AccessibilityEvent.TYPE_VIEW_ACCESSIBILITY_FOCUS_CLEARED:
    		return "View Accessibility Focus Cleared";
    	case AccessibilityEvent.TYPE_VIEW_ACCESSIBILITY_FOCUSED:
    		return "View Accessibility Focused";
    	case AccessibilityEvent.TYPE_VIEW_CLICKED:
    		return "View Clicked";
    	case AccessibilityEvent.TYPE_VIEW_FOCUSED:
    		return "View Focused";
    	case AccessibilityEvent.TYPE_VIEW_HOVER_ENTER:
    		return "View Hover Enter";
    	case AccessibilityEvent.TYPE_VIEW_HOVER_EXIT:
    		return "View Hover Exit";
    	case AccessibilityEvent.TYPE_VIEW_LONG_CLICKED:
    		return "View Long Clicked";
    	case AccessibilityEvent.TYPE_VIEW_SCROLLED:
    		return "View Scrolled";
    	case AccessibilityEvent.TYPE_VIEW_SELECTED:
    		return "View Selected";
    	case AccessibilityEvent.TYPE_VIEW_TEXT_CHANGED:
    		return "View Text Changed";
    	case AccessibilityEvent.TYPE_VIEW_TEXT_SELECTION_CHANGED:
    		return "View Text Selection Changed";
    	case AccessibilityEvent.TYPE_VIEW_TEXT_TRAVERSED_AT_MOVEMENT_GRANULARITY:
    		return "View Text Traversed at Movement Granularity";
    	case AccessibilityEvent.TYPE_WINDOW_CONTENT_CHANGED:
    		return "Window Content Changed";
    	case AccessibilityEvent.TYPE_WINDOW_STATE_CHANGED:
    		return "Window State Changed";
    	default:
    		return "Unknown (" + eventType + ")";
    	}
    }
    
    // Encodes the data about one accessibility node. This assumes an open JSON object.
    private void encodeNodeInfo(JSONStringer s, AccessibilityNodeInfo node, Object source, Rect parentBounds, Rect bounds) throws JSONException {
    	if (node == null) {
    		return;
    	}
    	
    	s.key("ClassName"); s.value(node.getClassName());
    	s.key("Source"); s.value(source == node); // FIXME: Not sure this works; is there isSameNode?
    	s.key("HasInputFocus"); s.value(node.isFocused());
    	s.key("HasAccessibilityFocus"); s.value(node.isAccessibilityFocused());
    	
    	s.key("ParentBounds");
    	s.object();
    	s.key("Top"); s.value(bounds.top - parentBounds.top);
    	s.key("Left"); s.value(bounds.left - parentBounds.left);
    	s.key("Width"); s.value(bounds.width());
    	s.key("Height"); s.value(bounds.height());
    	s.endObject();
    }
    
    // Recursively encodes a subtree of accessibility nodes
    private void encodeSubtree(JSONStringer s, AccessibilityNodeInfo node, Object source, Rect parentBounds) throws JSONException {
    	if (node == null) {
    		s.value(null);
    		return;
    	}
    	
    	if (node.getWindowId() != windowId) {
    		// FIXME: not sure if this check works; neither is this surfaced in the UI
    		s.value("Disconnected node");
    		return;
    	}
    	
    	s.object();
    	
    	Rect bounds = new Rect();
    	node.getBoundsInScreen(bounds);
    	encodeNodeInfo(s, node, source, parentBounds, bounds);

    	s.key("Children");
    	s.array();
    	int childCount = node.getChildCount();
    	for (int i = 0; i < childCount; i++) {
    		AccessibilityNodeInfo child = node.getChild(i);
    		if (child == null) {
    			break;
    		}
    		encodeSubtree(s, child, source, bounds);
    		child.recycle();
    	}
    	s.endArray();
    	
    	s.endObject();
    }
    
    @Override
    public void onAccessibilityEvent(AccessibilityEvent event) {
    	windowId = event.getWindowId();
   
    	try {
    		JSONStringer s = new JSONStringer();
	    	s.object();
	    	
	    	s.key("EventType"); s.value(eventTypeToString(event.getEventType()));
	    	s.key("EventTime"); s.value(event.getEventTime());

	    	s.key("Root");
	    	AccessibilityNodeInfo source = event.getSource();
	    	AccessibilityNodeInfo root = getRootInActiveWindow();
	    	encodeSubtree(s, root, source, new Rect());
	    	root.recycle();
	    	if (source != root && source != null) {
	    		source.recycle();
	    	}
	    	
	    	s.endObject();
	    	
	    	send(s);
	 	} catch (JSONException e) {
	 		Log.e(LOG_TAG, e.toString());
	 	}
    }
    
    private void send(final String message) {
       	synchronized (lock) {
        	if (writer == null) {
        		return;
        	}

    	    messages.add(message);
    		lock.notifyAll();
    	}
    }
    
    private void send(JSONStringer s) {
    	send(s.toString() + "\n");
    }
    
    private void close() {
    	new AsyncTask<Void,Void,Void>() {
    		@Override
    		public Void doInBackground(Void... unused) {
    			synchronized (lock) {
	    			try {
	    				serverSocket.close();
	    			} catch (IOException e) {
	    				Log.e(LOG_TAG, e.toString());
	    			}
	    		   	try {
	    	    		writer.close();
	    	    	} catch (IOException e) {
	    	    		Log.e(LOG_TAG, e.toString());
	    	    	}
	    		   	serverSocket = null;
	    	    	writer = null;
	    	    	messages.clear();
    			}
    	    	return null;
    		}
    	}.execute();
    }
    
    private int windowId;
    private Object lock;
    private List<String> messages;
    private ServerSocket serverSocket;
    private Writer writer;
}
